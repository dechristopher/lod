package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
)

var (
	// Version of LOD
	Version = "0.0.1"

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
	Cache       Cache    `json:"cache" toml:"cache"`               // cache configuration for this proxy instance
}

// Cache configuration for a Proxy instance
type Cache struct {
	MemCap   int `json:"mem_cap" toml:"mem_cap"`     // maximum number of tiles to store in the in-memory LRU cache
	MemPrune int `json:"mem_prune" toml:"mem_prune"` // number tiles to prune when we hit the MemCap
	MemTTL   int `json:"mem_ttl" toml:"mem_ttl"`     // in-memory cache TTL in seconds
	// Note: our redis cache does not have a max cap on tiles. It will grow unbounded, so
	// you must use a TTL to avoid capping out your cluster if you have a large tile set.
	RedisTTL int `json:"redis_ttl" toml:"redis_ttl"` // redis tile cache TTL in seconds (or -1 for no TTL)
}

var defaultCache = Cache{
	MemCap:   1000,
	MemPrune: 50,
	MemTTL:   86400,
	RedisTTL: 604800,
}

var zeroCache = Cache{
	MemCap:   0,
	MemPrune: 0,
	MemTTL:   0,
	RedisTTL: 0,
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
	return validate(Cap)
}

// validate instance Capabilities for sanity and errors
func validate(c Capabilities) error {
	if c.Instance.Port == 0 {
		return errors.New(fmt.Sprintf("invalid port provided port=%d", c.Instance.Port))
	}

	// validate each provided proxy endpoint configuration
	for num, proxy := range c.Proxies {
		if err := validateProxy(num, proxy); err != nil {
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
func validateProxy(num int, proxy Proxy) error {
	if proxy.Name == "" {
		return errors.New(fmt.Sprintf("configured proxy #%d cannot have an empty name", num))
	}

	matched, err := regexp.Match("^[a-zA-Z0-9_]+$", []byte(proxy.Name))
	if err != nil {
		return errors.Wrapf(err, "configured proxy name error name=%s", proxy.Name)
	}
	if !matched {
		return errors.New(fmt.Sprintf("configured proxy name may only "+
			"contain alphanumerics and underscores name=%s", proxy.Name))
	}

	if !strings.Contains(proxy.TileURL, "{z}") {
		return errors.New(fmt.Sprintf("proxy tile endpoint URL does not contain "+
			"{z} placeholder proxy=%s url=%s", proxy.Name, proxy.TileURL))
	}

	if !strings.Contains(proxy.TileURL, "{x}") {
		return errors.New(fmt.Sprintf("proxy tile endpoint URL does not contain "+
			"{x} placeholder proxy=%s url=%s", proxy.Name, proxy.TileURL))
	}

	if !strings.Contains(proxy.TileURL, "{y}") {
		return errors.New(fmt.Sprintf("proxy tile endpoint URL does not contain "+
			"{y} placeholder proxy=%s url=%s", proxy.Name, proxy.TileURL))
	}

	return validateCache(proxy)
}

// validateCache will validate a proxy endpoint's cache configuration
func validateCache(proxy Proxy) error {
	if proxy.Cache.MemCap < 1 {
		return errors.New(fmt.Sprintf("proxy cache cannot have zero or negative "+
			"capacity proxy=%s", proxy.Name))
	}

	if proxy.Cache.MemPrune < 1 {
		return errors.New(fmt.Sprintf("proxy cache cannot have zero or negative "+
			"prune size proxy=%s", proxy.Name))
	}

	if proxy.Cache.RedisTTL < 1 {
		return errors.New(fmt.Sprintf("proxy cache cannot have zero or negative "+
			"redis TTL proxy=%s", proxy.Name))
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
