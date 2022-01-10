package cache

import (
	"context"
	"sync"

	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"

	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
)

// Caches configured for this instance
var Caches CachesMap = make(map[string]*Cache)

// cacheLock is used internally to prevent redundant
// concurrent initialization of cache instances
var cacheLock *sync.Mutex

// CachesMap is an alias type for the map of proxy name to its cache
type CachesMap map[string]*Cache

// Cache is a wrapper struct that operates a dual cache against the in-memory
// cache and Redis as a backing cache
type Cache struct {
	internal *bigcache.BigCache // pointer to internal cache instance
	external *redis.Client      // pointer to external Redis cache
	Proxy    *config.Proxy      // copy of the proxy configuration
}

// OneMB represents one megabyte worth of bytes
const OneMB = 1024 * 1024

// Get a cache instance by name
func Get(name string) *Cache {
	if Caches[name] == nil {
		if cacheLock == nil {
			cacheLock = &sync.Mutex{}
		}

		cacheLock.Lock()
		defer cacheLock.Unlock()
		// find and populate a new cache instance for the given name
		for _, proxy := range config.Get().Proxies {
			if proxy.Name == name {
				conf := bigcache.DefaultConfig(proxy.Cache.MemTTLDuration)
				conf.StatsEnabled = !env.IsProd()
				conf.MaxEntrySize = 1024 * 10 // 100KB
				conf.HardMaxCacheSize = OneMB * proxy.Cache.MemCap

				internal, err := bigcache.NewBigCache(conf)
				if err != nil {
					util.Error(str.CCache, str.ECacheCreate, err.Error())
					return nil
				}

				var external *redis.Client

				if proxy.Cache.RedisURL != "" {
					opts, err := redis.ParseURL(proxy.Cache.RedisURL)
					if err != nil {
						util.Error(str.CCache, str.ECacheCreate, err.Error())
						return nil
					}
					external = redis.NewClient(opts)

					_, err = external.Ping(context.Background()).Result()
					if err != nil {
						util.Error(str.CCache, str.ECacheCreate, err.Error())
						return nil
					}
				}

				util.DebugFlag("cache", str.CCache, str.DCacheUp, name)

				Caches[name] = &Cache{
					internal: internal,
					external: external,
					Proxy:    &proxy,
				}
				return Caches[name]
			}
		}
		// if this happens, there's an edge case somewhere
		util.Error(str.CCache, str.ECacheName, name)
	}

	return Caches[name]
}

// Fetch will attempt to grab a tile by key from any of the cache layers,
// populating higher layers of the cache if found.
func (c *Cache) Fetch(key string) *TilePacket {
	cachedTile, err := c.internal.Get(key)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			util.DebugFlag("cache", str.CCache, str.DCacheMiss, key)
		} else {
			util.Error(str.CCache, str.ECacheFetch, key, err.Error())
			return nil
		}
	}

	if cachedTile == nil && c.external != nil {
		// try fetching from redis if not present in internal cache
		redisTile := c.external.Get(context.Background(), key)
		if redisTile.Err() != nil {
			if redisTile.Err() == redis.Nil {
				// exit early if we don't have anything cached at any level
				util.DebugFlag("cache", str.CCache, str.DCacheMissExt, key)
				return nil
			}
			util.Error(str.CCache, str.ECacheFetch, key, err.Error())
			return nil
		}

		// squeeze out the bytes from the redis response
		cachedTile, err = redisTile.Bytes()
		if err != nil {
			util.Error(str.CCache, str.ECacheFetch, key, err.Error())
			return nil
		}

		// if TTL set, extend Redis TTL when we fetch a tile to prevent
		// key expiry for tiles that are fetched periodically
		if c.Proxy.Cache.RedisTTLDuration > 0 {
			go c.external.Expire(context.Background(), key, c.Proxy.Cache.RedisTTLDuration)
		}
	}

	if cachedTile == nil {
		// exit if we don't have anything cached at any level
		util.DebugFlag("cache", str.CCache, str.DCacheMissExt, key)
		return nil
	}

	// wrap bytes in TilePacket container
	tile := TilePacket(cachedTile)
	// ensure we've got valid tile protobuf bytes
	if len(tile) == 0 || !tile.Validate() {
		// exit early and wipe cache if we cached a bad value
		util.DebugFlag("cache", str.CCache, str.DCacheFail, key)
		err = c.Invalidate(key)
		if err != nil {
			util.Error(str.CCache, str.ECacheDelete, key, err.Error())
		}
		return nil
	}

	util.DebugFlag("cache", str.CCache, str.DCacheHit, key, len(tile))

	// extend internal cache TTL (keeping entry alive) by resetting the entry
	// this also sets internal cache entries if we find a tile in redis but not internally
	// TODO investigate alternative methods of preventing entry death
	go c.Set(key, cachedTile, true)

	return &tile
}

// EncodeSet will encode tile data into a TilePacket and then set the cache
// entry to the specified key
func (c *Cache) EncodeSet(key string, tileData []byte, headers map[string]string) {
	packet := c.Encode(key, tileData, headers)
	c.Set(key, packet)
}

// Set the tile in all cache levels with the configured TTLs
func (c *Cache) Set(key string, tile TilePacket, internalOnly ...bool) {
	util.DebugFlag("cache", str.CCache, str.DCacheSet, key, len(tile))
	if (len(internalOnly) == 0 || !internalOnly[0]) && c.external != nil {
		go func() {
			status := c.external.Set(context.Background(), key,
				tile.Raw(), c.Proxy.Cache.RedisTTLDuration)
			if status.Err() != nil {
				util.Error(str.CCache, str.ECacheSet, key, status.Err())
			}
		}()
	}
	err := c.internal.Set(key, tile)
	if err != nil {
		util.Error(str.CCache, str.ECacheSet, key, err.Error())
	}
}

// Invalidate a tile by key from all cache levels
func (c *Cache) Invalidate(key string) error {
	err := c.internal.Delete(key)
	if err != nil && err != bigcache.ErrEntryNotFound {
		return err
	}

	if c.external != nil {
		status := c.external.Del(context.Background(), key)
		if status.Err() != nil {
			return status.Err()
		}
	}

	return nil
}

// Flush the internal bigcache instance
func (c *Cache) Flush() error {
	return c.internal.Reset()
}

func (c *Cache) Stats() bigcache.Stats {
	return c.internal.Stats()
}
