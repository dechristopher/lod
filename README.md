# LOD (Levels of Detail)

LOD is a thin vector map tile proxy with in-memory caching and a slim authentication backend. It can pull from any vector tile server and will aggressively cache tiles in memory and store them permanently in a configured Redis cluster for faster fetching later.
