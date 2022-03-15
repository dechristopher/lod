package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/util"
)

type stats struct {
	Hits     float64 `json:"hits"`     // total cache hits this proxy has encountered
	Misses   float64 `json:"misses"`   // total cache misses this proxy has encountered
	Requests float64 `json:"requests"` // total requests this proxy has serviced
	HitRate  float64 `json:"hit_rate"` // overall hit rate, hits/total requests
	TPS      float64 `json:"tps"`      // average tiles per second served over the past minute
	Cache    fetch   `json:"cache"`    // cache fetch performance stats
	Upstream fetch   `json:"upstream"` // upstream fetch performance stats
}

type fetch struct {
	FetchAvg  float64 `json:"fetch_avg"`   // average upstream fetch time over the past minute
	Fetch75th float64 `json:"fetch_75_th"` // 75th percentile upstream fetch time over the past minute
	Fetch99th float64 `json:"fetch_99_th"` // 99th percentile upstream fetch time over the past minute
}

// Stats returns stats for a cache by name, or all caches
func Stats(ctx *fiber.Ctx) error {
	if ctx.Path() == "/admin/stats" {
		return ctx.JSON(map[string]string{
			"stats": "all",
		})
	}

	name := ctx.Params("name")
	if name == "" {
		// quit early if no name provided
		return ctx.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"status": "bad request, no proxy name provided",
		})
	}

	c := cache.Get(name)
	if c == nil {
		// 404 if no proxy found with given name
		return ctx.Status(fiber.StatusNotFound).JSON(map[string]string{
			"status": "no proxy configured with given name",
		})
	}

	hits := util.GetMetricValue(c.Metrics.CacheHits)
	misses := util.GetMetricValue(c.Metrics.CacheMisses)

	return ctx.JSON(stats{
		Hits:     hits,
		Misses:   misses,
		Requests: hits + misses,
		HitRate:  util.GetMetricValue(c.Metrics.HitRate),
		TPS:      0,
		Cache: fetch{
			FetchAvg:  0,
			Fetch75th: 0,
			Fetch99th: 0,
		},
		Upstream: fetch{
			FetchAvg:  0,
			Fetch75th: 0,
			Fetch99th: 0,
		},
	})
}
