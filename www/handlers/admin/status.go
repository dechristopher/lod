package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/env"
	"github.com/dechristopher/lod/util"
)

type statusResponse struct {
	Version     string  `json:"v"`      // current lio version
	Environment env.Env `json:"env"`    // configured environment
	Uptime      float64 `json:"uptime"` // uptime in seconds
	BootTime    int64   `json:"boot"`   // time started, unix timestamp
}

// Status returns a JSON object with status info
func Status(c *fiber.Ctx) error {
	return c.JSON(statusResponse{
		Version:     config.Version,
		Environment: env.GetEnv(),
		Uptime:      util.TimeSinceBoot().Seconds(),
		BootTime:    util.BootTime.UnixMilli(),
	})
}
