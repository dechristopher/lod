package instance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/env"

	"github.com/tile-fund/lod/util"
)

type status struct {
	Version     string  `json:"v"`      // current lio version
	Environment env.Env `json:"env"`    // configured environment
	Uptime      float64 `json:"uptime"` // uptime in seconds
	BootTime    int64   `json:"boot"`   // time started, unix timestamp
}

// Status returns a JSON object with status info
func Status(c *fiber.Ctx) error {
	return c.JSON(status{
		Version:     config.Version,
		Environment: env.GetEnv(),
		Uptime:      util.TimeSinceBoot().Seconds(),
		BootTime:    util.BootTime.UnixMilli(),
	})
}
