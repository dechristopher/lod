# LOD (Level of Detail)

LOD is a thin vector map tile proxy with in-memory caching and a slim authentication backend. It can pull from any
vector tile server and will aggressively cache tiles in memory and store them permanently in a configured Redis cluster
for faster fetching later.

## Core Principles and Features

- Lightweight, parallel, and non-blocking
- Multi-level caching
    - In-memory, tunable LRU cache as first level
    - Redis cluster with configurable TTL as second level
- Internal stats tracking for top-N tiles requested
- Supports multiple configured tileserver proxies
  - Separate authentication (bearer tokens and CORS)
  - Separate stacks tracking and caches
