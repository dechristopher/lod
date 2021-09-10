package env

import "os"

// Env constants
type Env string

const (
	// Prod is the production environment
	Prod Env = "prod"
	// Dev is the dev environment
	Dev Env = "dev"
)

// GetEnv returns the current environment
func GetEnv() Env {
	if os.Getenv("DEPLOY") == "prod" {
		return Prod
	}
	return Dev
}

// IsProd returns true if the current deployed environment is production
func IsProd() bool {
	return GetEnv() == Prod
}

// IsDev returns true if the current deployed environment is not production
func IsDev() bool {
	return GetEnv() == Dev
}
