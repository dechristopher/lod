# [LOD](https://tile.fund/lod)
[![Go Report Card](https://goreportcard.com/badge/github.com/tile-fund/lod)](https://goreportcard.com/report/github.com/tile-fund/lod)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://raw.githubusercontent.com/tile-fund/lod/master/LICENSE)

LOD (Levels of Detail) is a thin vector map tile proxy with in-memory caching
and a slim authentication backend. It can pull from any vector tile server and
will aggressively cache tiles in memory, also storing them in a configured Redis
cluster for faster fetching later.

LOD is written in Go 1.17 using [fiber](https://github.com/gofiber/fiber).
Go templates are used for templating. Internal caching logic is built upon the
[ccache](https://github.com/karlseguin/ccache) library by
[karlseguin](https://github.com/karlseguin).

## Core Principles and Features

- Lightweight, parallel, and non-blocking
- Multi-level caching
    - In-memory, tunable LRU cache as first level
    - Redis cluster with configurable TTL as second level
- Internal stats tracking for top-N tiles requested
- Supports multiple configured tileserver proxies
  - Separate authentication (bearer tokens and CORS)
  - Separate stats tracking and caches

## License

LOD is licensed under the GNU Affero General Public License 3 or any later
version at your choice. See COPYING for details.
