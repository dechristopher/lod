package util

import (
	"net/url"
	"time"
)

// IsUrl returns true if the provided string is a valid URL with
// scheme and host set properly
func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// MilliTime returns the current millisecond time
func MilliTime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// TimeSinceBoot returns the time elapsed since process boot
func TimeSinceBoot() time.Duration {
	return time.Since(BootTime)
}
