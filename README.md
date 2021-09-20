# [LOD](https://tile.fund/lod) [![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://raw.githubusercontent.com/tile-fund/lod/master/LICENSE) ![Build & Test](https://github.com/tile-fund/lod/actions/workflows/build.yml/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/tile-fund/lod)](https://goreportcard.com/report/github.com/tile-fund/lod)

LOD (Levels of Detail) is a thin map tile proxy with in-memory caching and a 
slim authentication backend. It will sit in front of any tile server and will 
aggressively cache tiles in memory, optionally storing them in a configured
Redis cluster for faster fetching later. LOD is cluster-aware and uses Redis
message queueing for intra-cluster communication when multiple instances are
deployed together.

LOD is written in Go 1.17 using [fiber](https://github.com/gofiber/fiber). TOML
is used for configuration. Go templates are used for templating. Internal 
caching logic is built upon the [ccache](https://github.com/karlseguin/ccache)
library by [karlseguin](https://github.com/karlseguin).

## Core Principles

- Lightweight, parallel, and non-blocking
- Tileserver agnostic (Tegola, flat file NGINX, etc.)
- Tile format and content agnostic
  - Vector ([Mapbox Vector Tiles](https://github.com/mapbox/vector-tile-spec) 
    or [other vector formats](https://wiki.openstreetmap.org/wiki/Vector_tiles))
  - Raster (PNG/JPG/TIFF)
  - And [more](https://wiki.openstreetmap.org/wiki/Tiles)...
- Supports [XYZ (Slippy)](https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames)
  and [TMS](https://wiki.openstreetmap.org/wiki/TMS) tile indexing schemes

## v1.0.0 Feature Roadmap

- [ ] Multi-level caching
    - [X] In-memory, tunable LRU cache as first level
    - [ ] Redis cluster with configurable TTL as second level
- [X] Configurable header proxying and deletion
  - [X] `Content-Type` and `Content-Encoding` are added by default
- [ ] Internal stats tracking
  - [ ] Hits, misses, hit-rate
  - [ ] Top-N tiles requested
  - [ ] Tiles per second (load averages)
  - [ ] Tile upstream fetch times (avg, 75th, 99th)
  - [ ] Expose Prometheus endpoint
- [ ] Supports multiple configured tileserver proxies
  - [X] Separate authentication (bearer tokens and CORS)
  - [X] Separate internal caches
  - [ ] Allow query parameters for tile URLs
    - [ ] Add to cache key for separate caching (osm/4/5/6/osm_id=19)
  - [ ] Separate stats tracking
- [ ] Administrative endpoints
  - [ ] Reload the instance configuration
  - [ ] Invalidate the instance caches
  - [ ] Invalidate a given tile and re-prime it
  - [ ] Iteratively invalidate all tiles under a given tile (all zoom levels)
  - [ ] Iteratively prime all tiles under a given tile
  - [ ] Cluster-wide operations
    - [ ] Invalidate the instance caches across all instances
    - [ ] Invalidate a given tile and re-prime it across the cluster

## Sample Config
```toml
[instance]
port = 1337 # port to bind to

[[proxies]]
# name of this proxy, available at http://lod/{name}/{z}/{x}/{y}.pbf
name = "osm"
# url of the upstream tileserver
tile_url = "https://tile.example.com/osm/{z}/{x}/{y}.pbf" 
# comma-separated list of allowed CORS origins
cors_origins = "https://example.com"
# auth bearer token to require for requests to upstream tileserver
access_token = "MyTilesArePrivate" 
# headers to pull and cache from the tileserver response
add_headers = [ "X-We-Want-This", "X-This-One-Too" ] 
# headers to delete from the tileserver response
del_headers = [ "X-Get-Rid-Of-Me" ]

[proxies.cache]
mem_cap = 5000    # max capacity of in-memory cache
mem_ttl = 3600    # in-memory cache TTL in seconds
redis_ttl = 86400 # redis tile cache TTL in seconds (or -1 for no TTL)

# Supports many configured proxy instances for caching multiple tileservers
[[proxies]]
name = "another"
# etc.
```

## License

LOD is licensed under the GNU Affero General Public License 3 or any later
version at your choice. See COPYING for details.

## More Tile Resources
- https://wiki.openstreetmap.org/wiki/Category:Tiles_and_tiling
- https://wiki.openstreetmap.org/wiki/Tile_servers
- https://github.com/mapbox/awesome-vector-tiles
- https://docs.mapbox.com/vector-tiles/reference/
