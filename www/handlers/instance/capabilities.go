package instance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/config"
)

// Capabilities returns the JSON formatted output of the current
// instance configuration (capabilities)
func Capabilities(c *fiber.Ctx) error {
	return c.JSON(config.Cap)
}
