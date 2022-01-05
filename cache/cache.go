package cache

import (
	"time"

	"github.com/allegro/bigcache/v3"
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
	internal *bigcache.BigCache // pointer to internal cache instance
	proxy    config.Proxy       // copy of the proxy configuration
}

// Get a cache instance by name
func (cm CachesMap) Get(name string) *Cache {
	if cm[name] == nil {
		// find and populate a new cache instance for the given name
		for _, proxy := range config.Cap.Proxies {
			if proxy.Name == name {
				conf := bigcache.DefaultConfig(time.Duration(proxy.Cache.MemTTL) * time.Second)
				conf.StatsEnabled = true
				internal, err := bigcache.NewBigCache(conf)

				if err != nil {
					util.Error(str.CCache, str.ECacheCreate, err.Error())
					return nil
				}

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
func (c *Cache) Fetch(key string) *TilePacket {
	cachedTile, err := c.internal.Get(key)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			// exit early if we don't have anything cached
			go util.Debug(str.CCache, str.DCacheMiss, key)
			return nil
		}
		go util.Error(str.CCache, str.ECacheFetch, key, err.Error())
		return nil
	}

	if cachedTile == nil {
		// exit early if we don't have anything cached
		go util.Debug(str.CCache, str.DCacheMiss, key)
		return nil
	}

	// ensure we've got valid tile protobuf bytes
	tile := TilePacket(cachedTile)
	if len(tile) == 0 {
		// exit early and wipe cache if we cached a bad value
		go util.Debug(str.CCache, str.DCacheFail, key)
		err = c.internal.Delete(key)
		if err != nil {
			go util.Error(str.CCache, str.ECacheDelete, key, err.Error())
		}
		return nil
	}

	go util.Debug(str.CCache, str.DCacheHit, key, len(tile))

	// extend TTL (keeping entry alive) by resetting the entry
	// TODO investigate alternative methods of preventing entry death
	go c.Set(key, cachedTile)

	return &tile
}

// EncodeSet will encode tile data into a TilePacket and then set the cache
// entry to the specified key
func (c *Cache) EncodeSet(key string, tileData []byte, headers map[string]string) {
	packet := c.Encode(key, tileData, headers)
	c.Set(key, packet)
}

// Set the tile in all cache levels with the configured TTLs
func (c *Cache) Set(key string, tile TilePacket) {
	// go redis.Set(key, tile, etc)
	util.Debug(str.CCache, str.DCacheSet, key, len(tile))
	err := c.internal.Set(key, tile)
	if err != nil {
		util.Error(str.CCache, str.ECacheSet, key, err.Error())
	}
}

// Flush the internal bigcache instance
func (c *Cache) Flush() error {
	return c.internal.Reset()
}
