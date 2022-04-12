package cache

import (
	"context"
	"crypto/tls"
	"os"
	"strconv"

	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/env"
	"github.com/dechristopher/lod/packet"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/util"
)

var Subsystem = "cache"

// CachesMap is an alias type for the map of proxy name to its cache
type CachesMap map[string]*Cache

// Caches configured for this instance
var Caches = make(CachesMap)

// Cache is a wrapper struct that operates a dual cache against the in-memory
// cache and Redis as a backing cache
type Cache struct {
	internal *bigcache.BigCache // pointer to internal cache instance
	external *redis.Client      // pointer to external Redis cache
	Proxy    *config.Proxy      // a reference to the proxy's configuration
	Metrics  *Metrics           // metrics container instance
}

// Metrics for the cache instance
type Metrics struct {
	CacheHits   prometheus.Counter     // cache hits
	CacheMisses prometheus.Counter     // cache misses
	HitRate     prometheus.CounterFunc // cache hit rate
}

// OneMB represents one megabyte worth of bytes
const OneMB = 1024 * 1024

// Init cache instances for all configured proxies
func Init() error {
	// build out initial proxy instances
	for _, proxy := range config.Get().Proxies {
		if Caches[proxy.Name] == nil {
			err := BuildInstance(proxy.Name)
			if err != nil {
				return ErrBuildInstance{
					Name: proxy.Name,
					Err:  err,
				}
			}
		}
	}

	// ensure old proxies under different names aren't kept around
	WipeOldCaches()
	return nil
}

// Get a cache instance by name
func Get(name string) *Cache {
	return Caches[name]
}

// WipeOldCaches deletes old cache instances from the Caches map
// Usually called after a config read/reload
func WipeOldCaches() {
	proxies := config.Get().Proxies

	for cacheName := range Caches {
		present := false

		// ensure proxy is present in current config
		for _, proxy := range proxies {
			if cacheName == proxy.Name {
				present = true
				break
			}
		}

		// if present, don't delete
		if present {
			continue
		}

		// delete old cache if not present in current config
		util.Info(str.CCache, str.MOldCacheDeleted, cacheName)
		delete(Caches, cacheName)
	}
}

// BuildInstance will build a cache instance by name
func BuildInstance(name string) error {
	// find and populate a new cache instance for the given name
	for _, proxy := range config.Get().Proxies {
		if proxy.Name == name {
			var internal *bigcache.BigCache
			var external *redis.Client
			var err error

			if proxy.Cache.MemEnabled {
				internal, err = initInternal(proxy)
				if err != nil {
					return ErrInitInternalCache{
						Name: proxy.Name,
						Err:  err,
					}
				}
			}

			if proxy.Cache.RedisEnabled {
				external, err = initExternal(proxy)
				if err != nil {
					return ErrInitExternalCache{
						Name: proxy.Name,
						Err:  err,
					}
				}
			}

			// initialize metrics for this cache instance
			metrics := initMetrics(proxy)

			util.DebugFlag("cache", str.CCache, str.DCacheUp, name)

			Caches[name] = &Cache{
				internal: internal,
				external: external,
				Proxy:    &proxy,
				Metrics:  metrics,
			}

			return nil
		}
	}

	return nil
}

// initInternal initializes an in-memory cache instance from proxy configuration
func initInternal(proxy config.Proxy) (*bigcache.BigCache, error) {
	maxEntrySize := 3

	// allow override of MaxEntrySize via env var
	if max, present := os.LookupEnv("MAX_ENTRY_SIZE"); present {
		if maxInt, err := strconv.Atoi(max); err == nil {
			maxEntrySize = maxInt
		} else {
			util.Error(str.CCache, str.ECacheEntry, max)
		}
	}

	conf := bigcache.DefaultConfig(proxy.Cache.MemTTLDuration)
	conf.StatsEnabled = !env.IsProd()
	conf.MaxEntrySize = OneMB * 3
	conf.MaxEntrySize = OneMB * maxEntrySize
	conf.HardMaxCacheSize = proxy.Cache.MemCap

	return bigcache.NewBigCache(conf)
}

// initExternal initializes an external cache instance from proxy configuration
func initExternal(proxy config.Proxy) (*redis.Client, error) {
	if proxy.Cache.RedisTLS {
		proxy.Cache.RedisOpts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	// create new Redis client using opts parsed from config
	external := redis.NewClient(proxy.Cache.RedisOpts)

	// ping Redis to verify connectivity
	_, err := external.Ping(context.Background()).Result()

	return external, err
}

// initMetrics for the given proxy configuration
func initMetrics(proxy config.Proxy) *Metrics {
	cacheHits := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: config.Namespace,
		Subsystem: Subsystem,
		Name:      "hit_total",
		ConstLabels: map[string]string{
			"proxy": proxy.Name,
		},
		Help: "The total number of cache hits",
	})

	cacheMisses := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: config.Namespace,
		Subsystem: Subsystem,
		Name:      "miss_total",
		ConstLabels: map[string]string{
			"proxy": proxy.Name,
		},
		Help: "The total number of cache misses",
	})

	hitRate := promauto.NewCounterFunc(prometheus.CounterOpts{
		Namespace: config.Namespace,
		Subsystem: Subsystem,
		Name:      "hit_rate",
		ConstLabels: map[string]string{
			"proxy": proxy.Name,
		},
		Help: "The rate of hits to misses",
	}, func() float64 {
		hits := util.GetMetricValue(cacheHits)
		misses := util.GetMetricValue(cacheMisses)
		return hits / (hits + misses)
	})

	return &Metrics{
		CacheHits:   cacheHits,
		CacheMisses: cacheMisses,
		HitRate:     hitRate,
	}
}

// Fetch will attempt to grab a tile by key from any of the cache layers,
// populating higher layers of the cache if found.
func (c *Cache) Fetch(key string, ctx *fiber.Ctx) *packet.TilePacket {
	var cachedTile []byte
	var err error
	var hit string

	// fetch from in-memory cache if enabled
	if c.Proxy.Cache.MemEnabled {
		cachedTile, err = c.internal.Get(key)
		if err != nil {
			if err == bigcache.ErrEntryNotFound {
				util.DebugFlag("cache", str.CCache, str.DCacheMiss, key)
			} else {
				util.Error(str.CCache, str.ECacheFetch, key, err.Error())
				return nil
			}
		}

		hit = " :hit-i"
	}

	// try fetching from redis if not present in internal cache
	if cachedTile == nil && c.Proxy.Cache.RedisEnabled {
		var redisTile *redis.StringCmd

		if c.Proxy.Cache.RedisTTLDuration > 0 {
			// if TTL set, extend Redis TTL when we fetch a tile to prevent
			// key expiry for tiles that are fetched periodically
			redisTile = c.external.GetEx(ctx.Context(), key, c.Proxy.Cache.RedisTTLDuration)
		} else {
			// get and persist the key, meaning no expiry
			redisTile = c.external.GetEx(ctx.Context(), key, 0)
		}

		if redisTile.Err() != nil {
			if redisTile.Err() == redis.Nil {
				// exit early if we don't have anything cached at any level
				c.Metrics.CacheMisses.Inc()
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

		hit = " :hit-e"
	}

	if cachedTile == nil {
		// exit if we don't have anything cached at any level
		c.Metrics.CacheMisses.Inc()
		util.DebugFlag("cache", str.CCache, str.DCacheMissExt, key)
		return nil
	}

	ctx.Locals("lod-cache", hit)
	c.Metrics.CacheHits.Inc()

	// wrap bytes in TilePacket container
	tile, err := packet.FromBytes(cachedTile, key)
	if err != nil {
		// exit early and wipe cache if we cached a bad value
		util.Error(str.CCache, str.ECacheFetch, key, err.Error())
		err = c.Invalidate(key, ctx.Context())
		if err != nil {
			util.Error(str.CCache, str.ECacheDelete, key, err.Error())
		}
		return nil
	}

	util.DebugFlag("cache", str.CCache, str.DCacheHit, key, tile.TileDataSize())

	// extend internal cache TTL (keeping entry alive) by resetting the entry
	// this also sets internal cache entries if we find a tile in redis but not internally
	// TODO investigate alternative methods of preventing entry death
	go c.Set(key, *tile, true)

	return tile
}

// EncodeSet will encode tile data into a TilePacket and then set the cache
// entry to the specified key
func (c *Cache) EncodeSet(key string, tileData []byte, headers map[string]string) {
	tilePacket := packet.Encode(tileData, headers)
	c.Set(key, tilePacket)
}

// Set the tile in all cache levels with the configured TTLs
func (c *Cache) Set(key string, tile packet.TilePacket, internalOnly ...bool) {
	util.DebugFlag("cache", str.CCache, str.DCacheSet, key, len(tile))

	// set in external cache if enabled and allowed
	if (len(internalOnly) == 0 || !internalOnly[0]) && c.Proxy.Cache.RedisEnabled {
		go func() {
			status := c.external.Set(context.Background(), key,
				tile.Raw(), c.Proxy.Cache.RedisTTLDuration)
			if status.Err() != nil {
				util.Error(str.CCache, str.ECacheSet, key, status.Err())
			}
		}()
	}

	// set in the in-memory cache if enabled
	if c.Proxy.Cache.MemEnabled {
		err := c.internal.Set(key, tile)
		if err != nil {
			util.Error(str.CCache, str.ECacheSet, key, err.Error())
		}
	}
}

// Invalidate a tile by key from all cache levels
func (c *Cache) Invalidate(key string, ctx context.Context) error {
	// invalidate from in-memory cache if enabled
	if c.Proxy.Cache.MemEnabled {
		err := c.internal.Delete(key)
		if err != nil && err != bigcache.ErrEntryNotFound {
			return err
		}
	}

	if c.Proxy.Cache.RedisEnabled {
		status := c.external.Del(ctx, key)
		if status.Err() != nil {
			return status.Err()
		}
	}

	return nil
}

// FlushInternal flushes the internal bigcache instance
func (c *Cache) FlushInternal() error {
	if c.Proxy.Cache.MemEnabled {
		return c.internal.Reset()
	}
	return nil
}

// StatsInternal returns stats about the internal bigcache instance
func (c *Cache) StatsInternal() bigcache.Stats {
	if c.Proxy.Cache.MemEnabled {
		return c.internal.Stats()
	}
	return bigcache.Stats{}
}
