package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/config"
)

// Capabilities returns the JSON formatted output of the current
// instance configuration (capabilities)
func Capabilities(c *fiber.Ctx) error {
	return c.JSON(config.Get())
}
