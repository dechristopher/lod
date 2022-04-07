package cache

import "fmt"

// ErrBuildInstance is an error struct for errors
// encountered during a proxy's cache initialization
type ErrBuildInstance struct {
	Name string
	Err  error
}

// Error returns the string representation of ErrBuildInstance
func (e ErrBuildInstance) Error() string {
	return fmt.Sprintf("config: failed to init caches for proxy '%s', got error %s", e.Name, e.Err.Error())
}

// ErrInitInternalCache is an error struct for errors
// encountered during the internal cache initialization
type ErrInitInternalCache struct {
	Name string
	Err  error
}

// Error returns the string representation of ErrInitInternalCache
func (e ErrInitInternalCache) Error() string {
	return fmt.Sprintf("cache: failed to init internal memory cache for '%s', got error %s", e.Name, e.Err.Error())
}

// ErrInitExternalCache is an error struct for errors
// encountered during the external cache initialization
type ErrInitExternalCache struct {
	Name string
	Err  error
}

// Error returns the string representation of ErrInitExternalCache
func (e ErrInitExternalCache) Error() string {
	return fmt.Sprintf("cache: failed to init external cache for '%s', got error %s", e.Name, e.Err.Error())
}
