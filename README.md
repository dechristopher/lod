<!--suppress HtmlDeprecatedAttribute -->

<h1 align="center"><a href="https://lod.tile.fund">LOD</a>: Levels of Detail</h1>
<p align="center">An intelligent map tile proxy cache for the edge.</p>

<p align="center">
  <a href="https://github.com/tile-fund/lod/releases/latest" style="text-decoration: none">
    <img src="https://img.shields.io/github/v/release/tile-fund/lod?style=flat-square" alt="Latest Release">
  </a>
  <a href="https://github.com/tile-fund/lod/stargazers" style="text-decoration: none">
    <img src="https://img.shields.io/github/stars/tile-fund/lod.svg?style=flat-square" alt="Stars">
  </a>
  <a href="https://github.com/tile-fund/lod/fork" style="text-decoration: none">
    <img src="https://img.shields.io/github/forks/tile-fund/lod.svg?style=flat-square" alt="Forks">
  </a>
  <a href="https://opensource.org/licenses/AGPL-3.0" style="text-decoration: none">
    <img src="https://img.shields.io/badge/license-AGPL%20v3-blue.svg?style=flat-square" alt="License: AGPL v3">
  </a>
  <br/>
  <a href="https://github.com/tile-fund/lod/releases" style="text-decoration: none">
    <img src="https://img.shields.io/badge/platforms-linux%20%7C%20macos%20%7C%20windows-informational?style=for-the-badge" alt="Downloads">
  </a>
  <br/>
  <a href="https://github.com/tile-fund/lod/actions/workflows/build.yml" style="text-decoration: none">
    <img src="https://img.shields.io/github/workflow/status/tile-fund/lod/build?style=flat-square" alt="Build Status">
  </a>
  <a href="https://lod.tile.fund" style="text-decoration: none">
    <img src="https://img.shields.io/badge/docs-here-success?style=flat-square" alt="Docs">
  </a>
  <a href="https://goreportcard.com/report/github.com/tile-fund/lod" style="text-decoration: none">
    <img src="https://img.shields.io/badge/go%20report-A+-success.svg?style=flat-square" alt="Go Report Card">
  </a>
  <br/>
  <a href="https://codecov.io/gh/tile-fund/lod">
    <img src="https://img.shields.io/codecov/c/gh/tile-fund/lod?color=magenta&logo=codecov&style=flat-square" alt="Coverage"/>
  </a>
</p>

LOD (Levels of Detail) is a thin map tile proxy with in-memory caching and a 
slim authentication backend. It will sit in front of any tile server and will 
aggressively cache tiles in memory, optionally storing them in a configured
Redis cluster for faster fetching later. LOD is cluster-aware and uses Redis
message queueing for intra-cluster communication when multiple instances are
deployed together.

LOD is written in Go 1.17 using [fiber](https://github.com/gofiber/fiber). TOML
is used for configuration. Go templates are used for templating. Internal 
in-memory caching is built upon the [bigcache](https://github.com/allegro/bigcache)
library by [allegro](https://github.com/allegro).

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

## v1.0 Feature Roadmap

- [ ] Multi-level caching
    - [X] In-memory, tunable LRU cache as first level
    - [X] Redis cluster with configurable TTL as second level
- [X] Configurable header proxying and deletion
  - [X] `Content-Type` and `Content-Encoding` added by default
- [ ] Internal stats tracking
  - [ ] Hits, misses, hit-rate
  - [ ] Tiles per second (load averages)
  - [ ] Tile upstream fetch times (avg, 75th, 99th)
  - [ ] Expose Prometheus endpoint
- [ ] Supports multiple configured tileserver proxies
  - [X] Separate authentication (bearer tokens and CORS)
  - [X] Separate internal cache instances per proxy
  - [X] Allow configurable query parameters for tile URLs
    - [X] Add to cache key for separate caching (osm/4/5/6/osm_id=19)
  - [ ] Separate stats tracking
- [ ] Administrative endpoints
  - [X] Reload the instance configuration
  - [X] Flush the instance caches
  - [ ] Invalidate a given tile and re-prime it
  - [X] Iteratively invalidate all tiles under a given tile (all zoom levels)
  - [ ] Iteratively prime all tiles under a given tile
  - [ ] Cluster-wide operations
    - [ ] Flush the instance caches across all instances
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
mem_cap = 100    # maximum capacity in MB of the in-memory cache
mem_ttl = 3600    # in-memory cache TTL in seconds
redis_ttl = 86400 # redis tile cache TTL in seconds (or 0 for no TTL)
redis_url = "redis://localhost:6379/0" # redis connection URL
key_template = "{z}/{x}/{y}" # cache key template string, supports parameter names


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
