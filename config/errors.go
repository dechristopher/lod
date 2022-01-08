package config

import "fmt"

// ErrInvalidPort is an error struct for invalid instance
// port, caught during the instance validation phase
type ErrInvalidPort struct {
	Port int
}

// Error returns the string representation of ErrInvalidPort
func (e ErrInvalidPort) Error() string {
	return fmt.Sprintf("config:instance invalid port '%d', valid ports are 1-65535", e.Port)
}

// ErrProxyNoName is an error struct for a proxy defined
// without a name, caught during the proxy param validation phase
type ErrProxyNoName struct {
	Number int
}

// Error returns the string representation of ErrProxyNoName
func (e ErrProxyNoName) Error() string {
	return fmt.Sprintf("config:proxy(#%d) defined without a name", e.Number)
}

// ErrProxyName is an error struct for a proxy defined with a name that caused
// an error during validation, caught during the proxy param validation phase
type ErrProxyName struct {
	Number int
	Err    error
}

// Error returns the string representation of ErrProxyName
func (e ErrProxyName) Error() string {
	return fmt.Sprintf("config:proxy(#%d) name caused error during validation: %s",
		e.Number, e.Err.Error())
}

// ErrProxyInvalidName is an error struct for a proxy defined with a name that contains
// illegal characters, caught during the proxy param validation phase
type ErrProxyInvalidName struct {
	Number    int
	ProxyName string
}

// Error returns the string representation of ErrProxyInvalidName
func (e ErrProxyInvalidName) Error() string {
	return fmt.Sprintf("config:proxy(#%d) name '%s' may only contain alphanumerics and underscores",
		e.Number, e.ProxyName)
}

// ErrMissingTileURLTemplate is an error struct for a proxy cache key template
// without a required parameter, caught during the proxy param validation phase
type ErrMissingTileURLTemplate struct {
	ProxyName string
	TileURL   string
	Parameter string
}

// Error returns the string representation of ErrMissingTileURLTemplate
func (e ErrMissingTileURLTemplate) Error() string {
	return fmt.Sprintf("config:proxy(%s) tile URL template '%s' missing required parameter %s",
		e.ProxyName, e.TileURL, e.Parameter)
}

// ErrInvalidMemCap is an error struct for invalid memory
// cache capacity, caught during the proxy validation phase
type ErrInvalidMemCap struct {
	ProxyName string
}

// Error returns the string representation of ErrInvalidMemCap
func (e ErrInvalidMemCap) Error() string {
	return fmt.Sprintf("config:proxy(%s):cache cannot have zero or negative capacity",
		e.ProxyName)
}

// ErrInvalidMemTTL is an error struct for invalid memory
// cache TTL, caught during the proxy cache validation phase
type ErrInvalidMemTTL struct {
	ProxyName string
	TTL       string
}

// Error returns the string representation of ErrInvalidMemTTL
func (e ErrInvalidMemTTL) Error() string {
	return fmt.Sprintf("config:proxy(%s):cache invalid memory TTL of '%s', "+
		"valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"",
		e.ProxyName, e.TTL)
}

// ErrInvalidRedisTTL is an error struct for invalid redis
// cache TTL, caught during the proxy cache validation phase
type ErrInvalidRedisTTL struct {
	ProxyName string
	TTL       string
}

// Error returns the string representation of ErrInvalidRedisTTL
func (e ErrInvalidRedisTTL) Error() string {
	return fmt.Sprintf("config:proxy(%s):cache invalid Redis TTL of '%s', "+
		"valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\"",
		e.ProxyName, e.TTL)
}

// ErrMissingCacheTemplate is an error struct for a proxy cache key template
// without a required parameter, caught during the proxy param validation phase
type ErrMissingCacheTemplate struct {
	ProxyName string
	Template  string
	Parameter string
}

// Error returns the string representation of ErrMissingCacheTemplate
func (e ErrMissingCacheTemplate) Error() string {
	return fmt.Sprintf("config:proxy(%s):cache key template '%s' missing required parameter %s",
		e.ProxyName, e.Template, e.Parameter)
}

// ErrParamNoName is an error struct for a proxy parameter
// without a name, caught during the proxy param validation phase
type ErrParamNoName struct {
	ProxyName string
	Number    int
}

// Error returns the string representation of ErrParamNoName
func (e ErrParamNoName) Error() string {
	return fmt.Sprintf("config:proxy(%s):params no name defined for parameter #%d",
		e.ProxyName, e.Number)
}

// ErrParamNameDuplicate is an error struct for duplicate proxy
// parameter names, caught during the proxy param validation phase
type ErrParamNameDuplicate struct {
	ProxyName string
	Parameter Param
}

// Error returns the string representation of ErrParamNameDuplicate
func (e ErrParamNameDuplicate) Error() string {
	return fmt.Sprintf("config:proxy(%s):params duplicate parameter with name '%s'",
		e.ProxyName, e.Parameter.Name)
}
