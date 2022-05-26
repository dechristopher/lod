package helpers

import (
	"fmt"
)

// ErrInvalidStatusCode is an error struct returned from ProcessResponse
// when a non-2xx status code is returned from the upstream tile server
type ErrInvalidStatusCode struct {
	StatusCode int
	CacheKey   string
}

// Error returns the string representation of ErrInvalidStatusCode
func (e ErrInvalidStatusCode) Error() string {
	return fmt.Sprintf("resp: got non-2xx status code: %d for tile at cache key '%s'",
		e.StatusCode, e.CacheKey)
}
