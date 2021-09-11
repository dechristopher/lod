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

	InfoFormat  = "INFO  [%s] %s\n"
	DebugFormat = "DEBUG [%s] %s\n"
	ErrorFormat = "ERROR [%s] %s\n"
)

// (C) Log caller names
const (
	CMain = "LOD"
	CLog  = "Log"
	CConf = "Conf"
)

// (E) Error messages
const (
	ELogFail        = "failed to log error=%s msg=%+v"
	EConfig         = "failed to read config file error=%s"
	EConfigNotFound = "config file not found at path: '%s'"
	EConfigPort     = "port override in env is invalid env=%s err=%s"
	ERead           = "read err: %s"
	EWrite          = "write err: meta=%+v error=%s"
)

// (U) User-facing error messages and codes
const ()

// (M) Standard info log messages
const (
	MDevMode  = "!! DEVELOPER MODE !!"
	MInit     = "LOD v%s - copyright 2021 Andrew DeChristopher <me@dchr.host>\n"
	MStarted  = "started in %s [env: %s][http: %d]"
	MShutdown = "shutting down"
	MExit     = "exit"
)

// (D) Debug log messages
const ()

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
