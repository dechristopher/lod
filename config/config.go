package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/env"
	"github.com/dechristopher/lod/tile"
	"github.com/dechristopher/lod/util"
)

var (
	// Version of LOD
	Version = ".dev"

	Namespace = "lod"

	// capabilities is a store for local instance Capabilities
	capabilities Capabilities

	// File is a reference to the config file path to read from
	File *string
)

// Capabilities of the LOD instance (the configuration)
type Capabilities struct {
	Version  string   `json:"version"`                  // version string shown when viewing capabilities endpoint
	Instance Instance `json:"instance" toml:"instance"` // instance configuration
	Proxies  []Proxy  `json:"proxies" toml:"proxies"`   // configured proxy instances
}

// Instance configuration for LOD
type Instance struct {
	Port           int    `json:"port" toml:"port"`                       // configured LOD port
	Environment    string `json:"environment"`                            // configured LOD environment
	AdminDisabled  bool   `json:"admin_disabled" toml:"admin_disabled"`   // whether the admin endpoints are disabled
	AdminToken     string `json:"-" toml:"admin_token"`                   // admin endpoint auth bearer token
	MetricsEnabled bool   `json:"metrics_enabled" toml:"metrics_enabled"` // whether metrics are enabled
}

// Proxy represents a configuration for a single endpoint proxy instance
type Proxy struct {
	Name             string   `json:"name" toml:"name"`                 // display name for this proxy
	TileURL          string   `json:"tile_url" toml:"tile_url"`         // templated tileserver URL that this instance will hit
	HasEndpointParam bool     `json:"has_endpoint_param"`               // internal variable to track whether this proxy has a dynamic endpoint configured
	CorsOrigins      string   `json:"cors_origins" toml:"cors_origins"` // allowed CORS origins, comma separated
	PullHeaders      []string `json:"pull_headers" toml:"pull_headers"` // additional headers to pull and cache from the tileserver
	DeleteHeaders    []string `json:"del_headers" toml:"del_headers"`   // headers to exclude from the tileserver response
	AddHeaders       []Header `json:"add_headers" toml:"add_headers"`   // headers to inject into upstream requests to tileserver
	AccessToken      string   `json:"-" toml:"access_token"`            // optional access token for incoming requests
	Params           []Param  `json:"params" toml:"params"`             // URL query parameter configurations for this instance
	Cache            Cache    `json:"cache" toml:"cache"`               // cache configuration for this proxy instance
}

// Header to inject in upstream request to tileserver
type Header struct {
	Name  string `json:"name" toml:"name"`   // header name
	Value string `json:"value" toml:"value"` // header value
}

// Param configuration for a proxy instance
type Param struct {
	Name    string `json:"name" toml:"name"`       // parameter name - exact match in URL and used as token value for cache key
	Default string `json:"default" toml:"default"` // default parameter value if none provided in URL
}

// Cache configuration for a Proxy instance
// Cache TTLs are set using Go's built-in time.ParseDuration
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
// For example: 1h, 300s, 1000ms, 2h35m, etc.
type Cache struct {
	MemEnabled     bool          `json:"mem_enabled" toml:"mem_enabled"`     // whether the in-memory cache is enabled
	MemCap         int           `json:"mem_cap" toml:"mem_cap"`             // maximum capacity in MB of the in-memory cache
	MemTTL         string        `json:"mem_ttl" toml:"mem_ttl"`             // in-memory cache TTL, ex: 1h, 30s, 1000ms, etc
	MemTTLDuration time.Duration `json:"-" toml:"-"`                         // parsed duration from MemTTL
	RedisEnabled   bool          `json:"redis_enabled" toml:"redis_enabled"` // whether the redis cache is enabled
	// Note: our redis cache does not have a max cap on tiles. It will grow unbounded, so
	// you must use a TTL to avoid capping out your cluster if you have a large tile set.
	RedisTTL         string        `json:"redis_ttl" toml:"redis_ttl"` // redis tile cache TTL, ex: 1h, 30s, 1000ms, etc
	RedisTTLDuration time.Duration `json:"-" toml:"-"`                 // parsed duration from RedisTTL
	// Example: redis://<user>:<password>@<host>:<port>/<db_number>
	RedisURL    string         `json:"-" toml:"redis_url"`               // full redis connection URL for parsing, SENSITIVE
	RedisTLS    bool           `json:"redis_tls" toml:"redis_tls"`       // whether to use TLS when connecting to the redis server
	RedisOpts   *redis.Options `json:"-" toml:"-"`                       // internal redis options, first parsed with config
	KeyTemplate string         `json:"key_template" toml:"key_template"` // cache key template, supports XYZ and URL parameters
}

var defaultCache = Cache{
	MemCap:      1000,
	MemTTL:      "24h",
	KeyTemplate: "{z}/{x}/{y}",
}

var zeroCache = Cache{
	MemCap:      0,
	MemTTL:      "",
	RedisTTL:    "",
	RedisURL:    "",
	KeyTemplate: "",
}

// Get returns a pointer to the global configuration
func Get() *Capabilities {
	return &capabilities
}

// Load config file into instance Capabilities
func Load() error {
	var newCapabilities Capabilities
	var configData []byte
	var err error

	if util.IsUrl(*File) {
		// read config file from URL if provided as a URL
		configData, err = readHttp()
	} else {
		// read config file from disk if provided as a local path
		configData, err = ioutil.ReadFile(*File)
		if err != nil {
			return err
		}
	}

	// expand environment variables present within raw config
	configData = []byte(os.ExpandEnv(string(configData)))

	// decode config file as TOML
	if _, err = toml.Decode(string(configData), &newCapabilities); err != nil {
		return err
	}

	// inject instance info to config for viewing in /capabilities
	newCapabilities.Instance.Environment = string(env.GetEnv())
	newCapabilities.Version = Version

	// validate configuration
	err = validateCapabilities(&newCapabilities)
	if err != nil {
		return err
	}

	// set capabilities after validation
	capabilities = newCapabilities

	// set default cache parameters if not provided
	setDefaults()

	return nil
}

// readHttp reads the config from the
func readHttp() ([]byte, error) {
	// fetch config from URL if valid
	resp, err := http.Get(*File)
	if err != nil {
		return nil, ErrConfigGetHTTP{
			URL: *File,
			Err: err,
		}
	}

	// only accept 200 status codes, reject all others
	if resp.StatusCode != 200 {
		return nil, ErrConfigGetHTTP{
			URL:    *File,
			Status: resp.StatusCode,
		}
	}

	// read all bytes from response body into configData
	configData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrConfigGetHTTP{
			URL: *File,
			Err: err,
		}
	}

	return configData, nil
}

// set default instance and cache properties for read configuration if not provided
func setDefaults() {
	if capabilities.Instance.Port == 0 {
		capabilities.Instance.Port = 3100
	}

	for i := range capabilities.Proxies {
		if capabilities.Proxies[i].Cache == zeroCache {
			capabilities.Proxies[i].Cache = defaultCache
		}

		if capabilities.Proxies[i].Cache.KeyTemplate == "" {
			capabilities.Proxies[i].Cache.KeyTemplate = defaultCache.KeyTemplate
		}

		if capabilities.Proxies[i].PullHeaders == nil {
			capabilities.Proxies[i].PullHeaders = make([]string, 0)
		}

		// Register default content headers
		capabilities.Proxies[i].registerHeader(fiber.HeaderContentType)
		capabilities.Proxies[i].registerHeader(fiber.HeaderContentEncoding)
	}
}

// validateCapabilities validates instance Capabilities for sanity and errors
func validateCapabilities(c *Capabilities) error {
	if c.Instance.Port < 1 || c.Instance.Port > 65535 {
		return ErrInvalidPort{Port: c.Instance.Port}
	}

	// validate each provided proxy endpoint configuration
	for num := range c.Proxies {
		if err := validateProxy(num, &c.Proxies[num]); err != nil {
			return err
		}
	}

	return nil
}

// registerHeader will add a header to the list of headers to pull through from
// the underlying configured tileserver
func (p *Proxy) registerHeader(header string) {
	found := false
	for _, key := range p.PullHeaders {
		if key == header {
			found = true
			break
		}
	}

	if !found {
		p.PullHeaders = append(p.PullHeaders, header)
	}
}

// DoPullHeaders will fill the given header map with configured headers
// extracted from proxied requests to store alongside tile data in TilePackets
func (p *Proxy) DoPullHeaders(resp *fiber.Response, headers map[string]string) {
	for _, header := range p.PullHeaders {
		headerValue := resp.Header.Peek(header)
		if len(headerValue) > 0 {
			headers[header] = string(headerValue)
		}
	}
}

// DoDeleteHeaders will strip headers from the response that are part of the
// DeleteHeaders list of headers to delete from the final response
func (p *Proxy) DoDeleteHeaders(c *fiber.Ctx) {
	for _, delHeader := range p.DeleteHeaders {
		c.Response().Header.Del(delHeader)
	}
}

// validateProxy will validate an individual proxy endpoint in the configuration
func validateProxy(num int, proxy *Proxy) error {
	if proxy.Name == "" {
		return ErrProxyNoName{Number: num + 1}
	}

	matched, err := regexp.Match("^[a-zA-Z0-9_-]+$", []byte(proxy.Name))
	if err != nil {
		panic(err)
	}

	if !matched {
		return ErrProxyInvalidName{
			Number:    num + 1,
			ProxyName: proxy.Name,
		}
	}

	if proxy.TileURL == "" {
		return ErrMissingTileURL{
			ProxyName: proxy.Name,
		}
	}

	// reflect presence of dynamic endpoint template in HasEndpointParam
	proxy.HasEndpointParam = strings.Contains(proxy.TileURL, tile.EndpointTemplate)

	if !strings.Contains(proxy.TileURL, "{z}") {
		return ErrMissingTileURLTemplate{
			ProxyName: proxy.Name,
			TileURL:   proxy.TileURL,
			Parameter: "{z}",
		}
	}

	if !strings.Contains(proxy.TileURL, "{x}") {
		return ErrMissingTileURLTemplate{
			ProxyName: proxy.Name,
			TileURL:   proxy.TileURL,
			Parameter: "{x}",
		}
	}

	if !strings.Contains(proxy.TileURL, "{y}") {
		return ErrMissingTileURLTemplate{
			ProxyName: proxy.Name,
			TileURL:   proxy.TileURL,
			Parameter: "{y}",
		}
	}

	// validate the proxy's cache configuration
	if errCache := validateCache(proxy); errCache != nil {
		return errCache
	}

	// validate the proxy's parameter configurations
	if errParams := validateParams(proxy); errParams != nil {
		return errParams
	}

	return nil
}

// validateCache will validate a proxy endpoint's cache configuration
func validateCache(proxy *Proxy) error {
	// ensure at least one cache is enabled
	if !proxy.Cache.MemEnabled && !proxy.Cache.RedisEnabled {
		return ErrNoCacheEnabled{
			ProxyName: proxy.Name,
		}
	}

	// validate internal cache configuration
	if err := validateInternalCache(proxy); err != nil {
		return err
	}

	// validate external cache configuration
	if err := validateExternalCache(proxy); err != nil {
		return err
	}

	if !strings.Contains(proxy.Cache.KeyTemplate, "{z}") {
		return ErrMissingCacheTemplate{
			ProxyName: proxy.Name,
			Template:  proxy.Cache.KeyTemplate,
			Parameter: "{z}",
		}
	}

	if !strings.Contains(proxy.Cache.KeyTemplate, "{x}") {
		return ErrMissingCacheTemplate{
			ProxyName: proxy.Name,
			Template:  proxy.Cache.KeyTemplate,
			Parameter: "{x}",
		}
	}

	if !strings.Contains(proxy.Cache.KeyTemplate, "{y}") {
		return ErrMissingCacheTemplate{
			ProxyName: proxy.Name,
			Template:  proxy.Cache.KeyTemplate,
			Parameter: "{y}",
		}
	}

	return nil
}

// validateInternalCache validates internal cache configuration
func validateInternalCache(proxy *Proxy) error {
	// parse and validate in-memory cache parameters if enabled
	if proxy.Cache.MemEnabled {
		if proxy.Cache.MemCap < 1 {
			return ErrInvalidMemCap{ProxyName: proxy.Name}
		}

		memTTL, err := time.ParseDuration(proxy.Cache.MemTTL)
		if err != nil {
			return ErrInvalidMemTTL{
				ProxyName: proxy.Name,
				TTL:       proxy.Cache.MemTTL,
			}
		}

		// reject negative or zero TTLs
		// since in-memory cache must expire
		if memTTL < 0 {
			return ErrInvalidMemTTL{
				ProxyName: proxy.Name,
				TTL:       proxy.Cache.MemTTL,
			}
		}

		proxy.Cache.MemTTLDuration = memTTL
	}

	return nil
}

// validateExternalCache validates external cache configuration
func validateExternalCache(proxy *Proxy) error {
	// parse and validate redis cache parameters if enabled
	if proxy.Cache.RedisEnabled {
		// validate URL
		var errParse error
		proxy.Cache.RedisOpts, errParse = redis.ParseURL(proxy.Cache.RedisURL)
		if errParse != nil {
			return ErrInvalidRedisURL{
				ProxyName: proxy.Name,
				URL:       proxy.Cache.RedisURL,
				Err:       errParse,
			}
		}

		// validate that TTL is sane
		if proxy.Cache.RedisTTL != "" {
			redisTTL, err := time.ParseDuration(proxy.Cache.RedisTTL)
			if err != nil {
				return ErrInvalidRedisTTL{
					ProxyName: proxy.Name,
					TTL:       proxy.Cache.RedisTTL,
				}
			}

			// reject negative TTLs
			if redisTTL < 0 {
				return ErrInvalidRedisTTL{
					ProxyName: proxy.Name,
					TTL:       proxy.Cache.RedisTTL,
				}
			}

			proxy.Cache.RedisTTLDuration = redisTTL
		} else {
			// set TTL duration to zero if none specified, meaning permanent persistence in Redis
			proxy.Cache.RedisTTLDuration = 0
		}
	}

	return nil
}

// validateParams ensures configured params have valid and non-overlapping names
func validateParams(proxy *Proxy) error {
	if len(proxy.Params) == 0 {
		return nil
	}

	// begin with reserved parameter names
	var usedNames = []string{"e", "z", "x", "y"}

	for i, param := range proxy.Params {
		if param.Name == "" {
			return ErrParamNoName{
				ProxyName: proxy.Name,
				Number:    i + 1,
			}
		}

		for _, name := range usedNames {
			if name == param.Name {
				return ErrParamNameDuplicate{
					ProxyName: proxy.Name,
					Parameter: param,
				}
			}
		}

		usedNames = append(usedNames, param.Name)
	}

	return nil
}

// GetPort returns the configured primary HTTP port
// or 3100 if none configured
func GetPort() int {
	return capabilities.Instance.Port
}

// GetListenPort returns the colon-formatted listen port
func GetListenPort() string {
	return fmt.Sprintf(":%d", GetPort())
}
