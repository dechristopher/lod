package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/env"

	"github.com/tile-fund/lod/www/middleware"
)

// Wire admin group and endpoint handlers
func Wire(r *fiber.App) {
	// admin handler group
	adminGroup := r.Group("/admin")

	// wire up all middleware components
	middleware.Wire(adminGroup)

	// enable auth middleware if admin token configured
	if config.Cap.Instance.AdminToken != "" {
		adminGroup.Use(authMiddleware)
	}

	// JSON service health / status handler
	adminGroup.Get("/status", Status)

	// capabilities endpoint shows configuration summary
	adminGroup.Get("/capabilities", Capabilities)

	// reload endpoint will reload capabilities configuration from config.File
	adminGroup.Get("/reload", ReloadCapabilities)

	// flush an entire proxy cache by name
	adminGroup.Get("/:name/flush/", Flush)

	// invalidate a given tile without re-priming
	adminGroup.Get("/:name/invalidate/:z/:x/:y", InvalidateTile)

	// invalidate and a given tile and all of its children up to a given max
	// maxZoom defaults to zoom level 12
	adminGroup.Get("/:name/invalidate/deep/:z/:x/:y", InvalidateTileDeep)
	adminGroup.Get("/:name/invalidate/deep/:z/:x/:y/:maxZoom", InvalidateTileDeep)

	// invalidate and prime a given tile
	adminGroup.Get("/:name/prime/:z/:x/:y", PrimeTile)

	// invalidate and prime a given tile and all of its children up to a given max
	// maxZoom defaults to zoom level 12
	adminGroup.Get("/:name/prime/deep/:z/:x/:y", PrimeTileDeep)
	adminGroup.Get("/:name/prime/deep/:z/:x/:y/:maxZoom", PrimeTileDeep)
}

// authMiddleware checks for valid bearer tokens
func authMiddleware(ctx *fiber.Ctx) error {
	if ctx.GetReqHeaders()["Authorization"] != "Bearer "+config.Cap.Instance.AdminToken {
		if env.IsDev() {
			// provide useful error messages when running in dev mode
			return ctx.Status(400).JSON(map[string]string{
				"status":  "error",
				"message": "failed to auth, invalid bearer token supplied",
			})
		}
		// otherwise, pretend nothing exists
		return ctx.Status(404).SendString("")
	}

	// continue normally if checks succeed
	return ctx.Next()
}
