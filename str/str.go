package str

// CPadding is the current max caller padding, dynamically increased
var CPadding = 0

// Flags, log formats and miscellaneous strings
const (
	FConfigFile      = "conf"
	FConfigFileUsage = "Path to TOML configuration file. Default: config.toml"

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
	CLog   = "Log"
	CConf  = "Config"
	CProxy = "Proxy"
	CCache = "Cache"
)

// (E) Error messages
const (
	ELogFail        = "failed to log error=%s msg=%+v"
	EConfig         = "failed to read config file error=%s"
	EConfigNotFound = "config file not found at path: '%s'"
	EConfigPort     = "port override in env is invalid env=%s err=%s"
	EWire           = "failed to wire up instance handlers error=%s"
	EBadRequest     = "failed to fetch URL parameters error=%s"
	ECacheName      = "failed to create cache, no proxy configured for name=%s"
	ERead           = "read err: %s"
	EWrite          = "write err: error=%s meta=%+v"
)

// (U) User-facing error messages and codes
const ()

// (M) Standard info log messages
const (
	MDevMode  = "!! DEVELOPER MODE !!"
	MInit     = "LOD v%s - copyright 2021 Andrew DeChristopher <me@dchr.host>\n"
	MStarted  = "started in %s [env: %s][http: %d]"
	MProxy    = "configured proxy [%s] -> %s"
	MShutdown = "shutting down"
	MExit     = "exit"
)

// (D) Debug log messages
const (
	DCacheSet  = "cache set key=%s len=%d"
	DCacheMiss = "cache miss key=%s"
	DCacheFail = "cache bad value key=%s"
	DCacheHit  = "cache hit key=%s len=%d"
)

// (T) Test messages
const ()

// Help message
const Help = `
Flags:
  --conf  Path to TOML configuration file. Default: config.toml
  --dev   Whether to enable developer mode. Default: false
  --debug Optional comma separated debug flags. Ex: foo,bar,baz
  --help  Shows this help menu.
Usage:
  lod [--conf config.toml] [--dev]
`
