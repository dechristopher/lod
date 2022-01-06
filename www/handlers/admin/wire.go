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

	// flush an entire proxy cache by name
	admin.Get("/:name/flush/", Flush)

	// invalidate a given tile without re-priming
	admin.Get("/:name/invalidate/:z/:x/:y", InvalidateTile)

	// invalidate and a given tile and all of its children up to a given max
	// maxZoom defaults to zoom level 12
	admin.Get("/:name/invalidate/deep/:z/:x/:y", InvalidateTileDeep)
	admin.Get("/:name/invalidate/deep/:z/:x/:y/:maxZoom", InvalidateTileDeep)

	// invalidate and prime a given tile
	admin.Get("/:name/prime/:z/:x/:y", PrimeTile)

	// invalidate and prime a given tile and all of its children up to a given max
	// maxZoom defaults to zoom level 12
	admin.Get("/:name/prime/deep/:z/:x/:y", PrimeTileDeep)
	admin.Get("/:name/prime/deep/:z/:x/:y/:maxZoom", PrimeTileDeep)
}
