package env

import "os"

// Env constants
type Env string

// IsDevFlag overrides the environment to set developer mode
var IsDevFlag *bool

const (
	// Prod is the production environment (default)
	Prod Env = "prod"
	// Dev is developer mode or dev environments
	Dev Env = "dev"
)

// GetEnv returns the current environment
func GetEnv() Env {
	// enable dev mode if DEPLOY key set to DEV or user provides "--dev" flag
	if os.Getenv("DEPLOY") == "DEV" || (IsDevFlag != nil && *IsDevFlag) {
		return Dev
	}
	return Prod
}

// IsProd returns true if the current deployed environment is production
func IsProd() bool {
	return GetEnv() == Prod
}

// IsDev returns true if the current deployed environment is not production
func IsDev() bool {
	return GetEnv() == Dev
}
