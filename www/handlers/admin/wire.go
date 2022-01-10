package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/www/middleware"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Wire admin group and endpoint handlers
func Wire(r *fiber.App) {
	// admin handler group
	adminGroup := r.Group("/admin")

	// wire up all middleware components
	middleware.Wire(adminGroup)

	// enable auth middleware if admin token configured
	if config.Cap.Instance.AdminToken != "" {
		adminGroup.Use(middleware.GenAuthMiddleware(config.Cap.Instance.AdminToken,
			middleware.Bearer, true))
	}

	if config.Cap.Instance.MetricsEnabled {
		// prometheus metrics endpoint
		p := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
		adminGroup.Get("/metrics/prometheus", func(c *fiber.Ctx) error {
			p(c.Context())
			return nil
		})
	}

	// JSON service health / status handler
	adminGroup.Get("/status", Status)

	// capabilities endpoint shows configuration summary
	adminGroup.Get("/capabilities", Capabilities)

	// reload endpoint will reload capabilities configuration from config.File
	adminGroup.Get("/reload", ReloadCapabilities)

	// flush all proxy caches
	adminGroup.Get("/flush", Flush)

	// flush an entire proxy cache by name
	adminGroup.Get("/:name/flush", Flush)

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
