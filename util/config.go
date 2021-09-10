package util

import (
	"fmt"
	"os"
	"time"
)

var (
	// Version of LOD
	Version = "0.0.1"

	// BootTime is set the instant everything comes online
	BootTime time.Time

	// DebugFlagPtr contains raw debug flags direct from STDIN
	DebugFlagPtr *string
	// DebugFlags holds all active, parsed debug flags
	DebugFlags []string
)

// IsDebugFlag returns true if a given debug flag is enabled in this instance
func IsDebugFlag(flag string) bool {
	for _, f := range DebugFlags {
		if f == flag {
			return true
		}
	}

	return false
}

// GetPort returns the configured primary HTTP port
func GetPort() string {
	return os.Getenv("PORT")
}

// GetListenPort returns the colon-formatted listen port
func GetListenPort() string {
	return fmt.Sprintf(":%s", GetPort())
}

// CorsOrigins returns the proper CORS origin configuration
//  or "*" if none configured
func CorsOrigins() string {
	origins, ok := os.LookupEnv("CORS")
	if !ok {
		return "*"
	}

	return origins
}
