# [LOD](https://tile.fund/lod)
[![Go Report Card](https://goreportcard.com/badge/github.com/tile-fund/lod)](https://goreportcard.com/report/github.com/tile-fund/lod)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://raw.githubusercontent.com/tile-fund/lod/master/LICENSE)

LOD (Levels of Detail) is a thin map tile proxy with in-memory caching and a 
slim authentication backend. It can pull from any tile server and will 
aggressively cache tiles in memory, optionally storing them in a configured
Redis cluster for faster fetching later.

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

## Features

- Multi-level caching
    - In-memory, tunable LRU cache as first level
    - Redis cluster with configurable TTL as second level
- Configurable header proxying and deletion
  - `Content-Type` and `Content-Encoding` are added by default
- Internal stats tracking for top-N tiles requested
- Supports multiple configured tileserver proxies
  - Separate authentication (bearer tokens and CORS)
  - Separate stats tracking and internal caches
- Administrative endpoints
  - Reload the instance configuration
  - Wipe the instance caches
  - Cluster-wide operations
    - Wipe the entire cache for a proxy on all instances
    - Wipe a given tile and re-prime it across the cluster
    - Iteratively wipe all tiles under a given tile (all zoom levels)
    - Iteratively prime all tiles under a given tile

## Sample Config
```toml
# Sample LOD config file

[instance]
port = 1337

[[proxies]]
name = "osm"
tile_url = "https://tile.example.com/osm/{z}/{x}/{y}.pbf"
cors_origins = "https://example.com"
access_token = "MyTilesArePrivate"
add_headers = [ "X-We-Want-This", "X-This-One-Too" ]
del_headers = [ "X-Get-Rid-Of-Me" ]

[proxies.cache]
mem_cap = 5000
mem_prune = 100
mem_ttl = 3600
redis_ttl = 86400
```

## License

LOD is licensed under the GNU Affero General Public License 3 or any later
version at your choice. See COPYING for details.

## More Tile Resources
- https://wiki.openstreetmap.org/wiki/Category:Tiles_and_tiling
- https://wiki.openstreetmap.org/wiki/Tile_servers
- https://github.com/mapbox/awesome-vector-tiles
- https://docs.mapbox.com/vector-tiles/reference/
