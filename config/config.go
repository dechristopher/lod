package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
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
	Name        string `json:"name" toml:"name"`                 // display name for this proxy
	URL         string `json:"proxy_url" toml:"proxy_url"`       // templated tileserver URL that this instance will hit
	CorsOrigins string `json:"cors_origins" toml:"cors_origins"` // allowed CORS origins, comma separated
	AccessToken string `json:"-" toml:"access_token"`            // optional access token for incoming requests
	Cache       Cache  `json:"cache" toml:"cache"`               // cache configuration for this proxy instance
}

// Cache configuration for a Proxy instance
type Cache struct {
	MemCap   int `json:"mem_cap" toml:"mem_cap"`     // maximum number of tiles to store in the in-memory LRU cache
	MemPrune int `json:"mem_prune" toml:"mem_prune"` // number tiles to prune when we hit the MemCap
	// Note: our redis cache does not have a max cap on tiles. It will grow unbounded, so
	// you must use a TTL to avoid capping out your cluster if you have a large tile set.
	RedisTTL int `json:"redis_ttl" toml:"redis_ttl"` // redis tile cache TTL in seconds (or -1 for no TTL)
}

// Read config file into Capabilities
func Read() error {
	configData, err := ioutil.ReadFile(*File)
	if err != nil {
		return err
	}

	_, err = toml.Decode(string(configData), &Cap)
	return err
}

// GetPort returns the configured primary HTTP port
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
//  or "*" if none configured
func CorsOrigins() string {
	origins, ok := os.LookupEnv("CORS")
	if !ok {
		return "*"
	}

	return origins
}
