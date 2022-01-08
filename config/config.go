package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
)

var (
	// Version of LOD
	Version = ".dev"

	// Cap is a store for local instance Capabilities
	Cap Capabilities

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
	Port        int    `json:"port" toml:"port"` // configured LOD port
	Environment string `json:"environment"`      // configured LOD environment
}

// Proxy represents a configuration for a single endpoint proxy instance
type Proxy struct {
	Name        string   `json:"name" toml:"name"`                 // display name for this proxy
	TileURL     string   `json:"tile_url" toml:"tile_url"`         // templated tileserver URL that this instance will hit
	CorsOrigins string   `json:"cors_origins" toml:"cors_origins"` // allowed CORS origins, comma separated
	AddHeaders  []string `json:"add_headers" toml:"add_headers"`   // additional headers to pull through from the tileserver
	DelHeaders  []string `json:"del_headers" toml:"del_headers"`   // headers to exclude from the tileserver response
	AccessToken string   `json:"-" toml:"access_token"`            // optional access token for incoming requests
	Params      []Param  `json:"params" toml:"params"`             // URL query parameter configurations for this instance
	Cache       Cache    `json:"cache" toml:"cache"`               // cache configuration for this proxy instance
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
	MemCap         int           `json:"mem_cap" toml:"mem_cap"` // maximum capacity in MB of the in-memory cache
	MemTTL         string        `json:"mem_ttl" toml:"mem_ttl"` // in-memory cache TTL, ex: 1h, 30s, 1000ms, etc
	MemTTLDuration time.Duration `json:"-" toml:"-"`             // parsed duration from MemTTL
	// Note: our redis cache does not have a max cap on tiles. It will grow unbounded, so
	// you must use a TTL to avoid capping out your cluster if you have a large tile set.
	RedisTTL         string        `json:"redis_ttl" toml:"redis_ttl"` // redis tile cache TTL, ex: 1h, 30s, 1000ms, etc
	RedisTTLDuration time.Duration `json:"-" toml:"-"`                 // parsed duration from RedisTTL
	// Example: redis://<user>:<password>@<host>:<port>/<db_number>
	RedisURL    string `json:"-" toml:"redis_url"`               // full redis connection URL for parsing, SENSITIVE
	KeyTemplate string `json:"key_template" toml:"key_template"` // cache key template, supports XYZ and URL parameters
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

// Read config file into Capabilities
func Read() error {
	configData, err := ioutil.ReadFile(*File)
	if err != nil {
		return err
	}

	if _, err = toml.Decode(string(configData), &Cap); err != nil {
		return err
	}

	// set default cache parameters if not provided
	for i := range Cap.Proxies {
		if Cap.Proxies[i].Cache == zeroCache {
			Cap.Proxies[i].Cache = defaultCache
		}

		if Cap.Proxies[i].Cache.KeyTemplate == "" {
			Cap.Proxies[i].Cache.KeyTemplate = defaultCache.KeyTemplate
		}

		if Cap.Proxies[i].AddHeaders == nil {
			Cap.Proxies[i].AddHeaders = make([]string, 0)
		}

		// Register default content headers
		Cap.Proxies[i].registerHeader("Content-Type")
		Cap.Proxies[i].registerHeader("Content-Encoding")
	}

	// inject instance info to config for viewing in /capabilities
	Cap.Instance.Environment = string(env.GetEnv())
	Cap.Version = Version

	// validate configuration
	return validate(&Cap)
}

// validate instance Capabilities for sanity and errors
func validate(c *Capabilities) error {
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
	for _, key := range p.AddHeaders {
		if key == header {
			found = true
			break
		}
	}

	if !found {
		p.AddHeaders = append(p.AddHeaders, header)
	}
}

// PopulateHeaders will fill the given header map with configured headers
func (p *Proxy) PopulateHeaders(c *fiber.Ctx, headers map[string]string) {
	for _, header := range p.AddHeaders {
		if len(c.Response().Header.Peek(header)) > 0 {
			headers[header] = string(c.Response().Header.Peek(header))
		}
	}
}

// DeleteHeaders will strip headers from the response that are part of the
// DelHeaders list of headers to delete from the final response
func (p *Proxy) DeleteHeaders(c *fiber.Ctx) {
	for _, delHeader := range p.DelHeaders {
		if len(c.Response().Header.Peek(delHeader)) > 0 {
			c.Response().Header.Del(delHeader)
		}
	}
}

// validateProxy will validate an individual proxy endpoint in the configuration
func validateProxy(num int, proxy *Proxy) error {
	if proxy.Name == "" {
		return ErrProxyNoName{Number: num + 1}
	}

	matched, err := regexp.Match("^[a-zA-Z0-9_]+$", []byte(proxy.Name))
	if err != nil {
		return ErrProxyName{
			Number: num + 1,
			Err:    err,
		}
	}

	if !matched {
		return ErrProxyInvalidName{
			Number:    num + 1,
			ProxyName: proxy.Name,
		}
	}

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

	// validate the proxy's parameter configurations
	if errParams := validateParams(proxy); errParams != nil {
		return errParams
	}

	// validate the proxy's cache configuration
	if errCache := validateCache(proxy); errCache != nil {
		return errCache
	}

	return nil
}

// validateCache will validate a proxy endpoint's cache configuration
func validateCache(proxy *Proxy) error {
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

	proxy.Cache.MemTTLDuration = memTTL

	redisTTL, err := time.ParseDuration(proxy.Cache.RedisTTL)
	if err != nil {
		return ErrInvalidRedisTTL{
			ProxyName: proxy.Name,
			TTL:       proxy.Cache.MemTTL,
		}
	}

	proxy.Cache.RedisTTLDuration = redisTTL

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

// validateParams ensures configured params don't have valid and non-overlapping names
func validateParams(proxy *Proxy) error {
	if len(proxy.Params) == 0 {
		return nil
	}

	var usedNames []string

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
// or 1337 if none configured
func GetPort() int {
	portEnv, ok := os.LookupEnv("PORT")
	if !ok {
		if Cap.Instance.Port != 0 {
			return Cap.Instance.Port
		}
		return 1337
	}
	port, err := strconv.Atoi(portEnv)
	if err != nil {
		util.Error(str.CConf, str.EConfigPort, portEnv, err.Error())
	}
	return port
}

// GetListenPort returns the colon-formatted listen port
func GetListenPort() string {
	return fmt.Sprintf(":%d", GetPort())
}

// CorsOrigins returns the proper CORS origin configuration
// or "*" if none configured
func CorsOrigins(p Proxy) string {
	origins, ok := os.LookupEnv("CORS")
	if !ok {
		if p.CorsOrigins != "" {
			return p.CorsOrigins
		}
		return "*"
	}

	return origins
}
