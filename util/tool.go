package util

import (
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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

// GetMetricValue extracts the current metric value from the
// given collector using prometheus collection internals
func GetMetricValue(col prometheus.Collector) float64 {
	c := make(chan prometheus.Metric, 1) // 1 for metric with no vector
	col.Collect(c)                       // collect current metric value into the channel
	m := dto.Metric{}
	_ = (<-c).Write(&m) // read metric value from the channel
	return *m.Counter.Value
}
