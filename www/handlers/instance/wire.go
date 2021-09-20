package instance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/www/middleware"
)

// Wire instance group and endpoint handlers
func Wire(r *fiber.App) {
	// instance handler group
	instanceGroup := r.Group("/instance")

	// wire up all middleware components
	middleware.Wire(instanceGroup)

	// capabilities endpoint shows configuration summary
	instanceGroup.Get("/capabilities", Capabilities)

	// JSON service health / status handler
	instanceGroup.Get("/status", Status)
}
