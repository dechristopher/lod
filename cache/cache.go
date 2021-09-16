package cache

import (
	"time"

	"github.com/karlseguin/ccache/v2"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
)

// Caches configured for this instance
var Caches CachesMap = make(map[string]*Cache)

// CachesMap is an alias type for the map of proxy name to its cache
type CachesMap map[string]*Cache

// Cache is a wrapper struct that operates a dual cache against the in-memory
// LRU cache and Redis as a backing cache
type Cache struct {
	internal *ccache.Cache // pointer to internal cache instance
	proxy    config.Proxy  // copy of the proxy configuration
}

// Tile is a wrapper struct that contains the tile PBF data and
// any additional HTTP headers provided by the tileserver
type Tile struct {
	Data    []byte
	Headers map[string]string
}

// Get a cache instance by name
func (cm CachesMap) Get(name string) *Cache {
	if cm[name] == nil {
		// find and populate a new cache instance for the given name
		for _, proxy := range config.Cap.Proxies {
			if proxy.Name == name {
				internal := ccache.New(ccache.Configure().
					MaxSize(int64(proxy.Cache.MemCap)).
					ItemsToPrune(uint32(proxy.Cache.MemPrune)),
				)
				cm[name] = &Cache{
					internal: internal,
					proxy:    proxy,
				}
				return cm[name]
			}
		}
		// if this happens, there's an edge case somewhere
		util.Error(str.CCache, str.ECacheName, name)
	}

	return cm[name]
}

// Fetch will attempt to grab a tile by key from any of the cache layers,
// populating higher layers of the cache if found.
func (c *Cache) Fetch(key string) *Tile {
	cachedTile := c.internal.Get(key)
	if cachedTile == nil {
		// exit early if we don't have anything cached
		go util.Debug(str.CCache, str.DCacheMiss, key)
		return nil
	}

	// ensure we've got valid tile protobuf bytes
	tile, ok := cachedTile.Value().(Tile)
	if !ok || len(tile.Data) == 0 {
		// exit early and wipe cache if we cached a bad value
		go util.Debug(str.CCache, str.DCacheFail, key)
		c.internal.Delete(key)
		return nil
	}

	go util.Debug(str.CCache, str.DCacheHit, key, len(tile.Data))

	// extend TTL if requested before expiry
	if !cachedTile.Expired() {
		cachedTile.Extend(time.Duration(c.proxy.Cache.MemTTL) * time.Second)
	}

	return &tile
}

// Set the tile in all cache levels with the configured TTLs
func (c *Cache) Set(key string, tile Tile) {
	c.internal.Set(key, tile, time.Duration(c.proxy.Cache.MemTTL)*time.Second)
	// go redis.Set(key, tile, etc)
	go util.Debug(str.CCache, str.DCacheSet, key, len(tile.Data))
}
