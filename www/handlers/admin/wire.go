package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/tile-fund/lod/www/middleware"
)

// Wire admin group and endpoint handlers
func Wire(r *fiber.App) {
	// admin handler group
	admin := r.Group("/admin")

	// wire up all middleware components
	middleware.Wire(admin)

	// reload endpoint will reload capabilities configuration from config.File
	admin.Get("/reload", ReloadCapabilities)
}
