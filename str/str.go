package str

// Flags, log formats and miscellaneous strings
const (
	FConfigFile      = "conf"
	FConfigFileUsage = "Path/URL to TOML configuration file. Default: config.toml"

	FDevMode      = "dev"
	FDevModeUsage = "Whether to enable developer mode. Default: false"

	FDebugFlags      = "debug"
	FDebugFlagsUsage = "Optional comma separated debug flags. Ex: foo,bar,baz"

	FHelp      = "help"
	FHelpUsage = "Shows this help menu."

	InfoFormat  = "INF [%s] %s\n"
	DebugFormat = "DBG [%s] %s\n"
	ErrorFormat = "ERR [%s] %s\n"
)

// (C) Log caller names
const (
	CMain  = "LOD"
	CLog   = "LOG"
	CConf  = "CFG"
	CProxy = "PRX"
	CCache = "CCH"
	CAdmin = "ADM"
)

// (E) Error messages
const (
	ELogFail            = "failed to log, error=%s msg=%+v"
	EConfig             = "failed to read config file: %s"
	EConfigNotFound     = "config file not found at path: '%s'"
	EConfigPort         = "port override in env is invalid, env=%s err=%s"
	EBadRequest         = "failed to fetch URL parameters, error=%s"
	ECacheCreate        = "failed to create internal cache, error=%s"
	ECacheName          = "failed to create cache, no proxy configured, name=%s"
	ECacheBuildKey      = "failed to build cache key err=%s"
	ECacheFetch         = "failed to fetch tile from cache, key=%s error=%s"
	ECacheDelete        = "failed to delete tile from cache, key=%s error=%s"
	ECacheSet           = "failed to set cache entry, key=%s error=%s"
	ECacheFlush         = "failed to flush cache, name=%s error=%s"
	EInvalidateTileDeep = "failed to invalidate tile %s with depth error=%s"
	EInvalidateTile     = "failed to invalidate tile %s error=%s"
	EPrimeTileDeep      = "failed to prime tile %s with depth error=%s"
	EPrimeTile          = "failed to prime tile %s error=%s"
	EWrite              = "write err: error=%s meta=%+v"
	EReload             = "failed to reload instance capabilities, error=%s"
	ERequest            = "generic uncaught error in request chain, ctx=%s error=%s"
)

// (U) User-facing error messages and codes
const ()

// (M) Standard info log messages
const (
	MDevMode            = "!! DEVELOPER MODE !!"
	MInit               = "LOD v%s - copyright 2021 Andrew DeChristopher <me@dchr.host>\n"
	MStarted            = "started in %s [env: %s][http: %d]"
	MProxy              = "configured proxy [mem: %t / redis: %t][%s] -> %s"
	MReload             = "reloaded instance capabilities"
	MInvalidateTile     = "invalidated tile %s with no depth (%d) (%d tiles)"
	MInvalidateTileDeep = "invalidated tile %s with depth %d (%d tiles)"
	MPrimeTile          = "primed tile %s with no depth (%d) (%d tiles)"
	MPrimeTileDeep      = "primed tile %s with depth %d (%d tiles)"
	MShutdown           = "shutting down"
	MExit               = "exit"
)

// (D) Debug log messages
const (
	DCacheUp        = "cache online name=%s"
	DCacheSet       = "cache set key=%s len=%d"
	DCacheMiss      = "cache internal miss key=%s"
	DCacheMissExt   = "cache external miss key=%s"
	DCacheFail      = "cache bad value key=%s"
	DCacheHit       = "cache hit key=%s len=%d"
	DInvalidateFail = "failed to invalidate tile %s, err=%s"
)

// (T) Test messages
const (
	TCacheEncodeHeaders = "retrieved headers length did not match input, got=%d expected=%d"
	TCacheBadHeaderData = "header data not properly encoded into tile packet"
	TCacheBadTileData   = "tile data not properly encoded into tile packet"
	TCacheBadValidation = "tile data corrupted, checksum failed"
	TCacheBadDecode     = "tile decode failed, error=%s"
)

// Help message
const Help = `
Flags:
  --conf  Path/URL to TOML configuration file. Default: config.toml
  --dev   Whether to enable developer mode. Default: false
  --debug Optional comma separated debug flags. Ex: foo,bar,baz
  --help  Shows this help menu.
Usage:
  lod [--conf config.toml] [--dev]
`
