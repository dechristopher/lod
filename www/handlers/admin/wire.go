package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/www/middleware"
)

// Wire admin group and endpoint handlers
func Wire(r *fiber.App) {
	// admin handler group
	adminGroup := r.Group("/admin")

	// wire up all middleware components
	middleware.Wire(adminGroup, nil)

	// enable auth middleware if admin token configured
	if config.Get().Instance.AdminToken != "" {
		adminGroup.Use(middleware.GenAuthMiddleware(config.Get().Instance.AdminToken,
			middleware.Bearer, true))
	}

	if config.Get().Instance.MetricsEnabled {
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

	// return stats for all caches
	adminGroup.Get("/stats", Stats)

	// flush the in-memory caches of all proxies
	adminGroup.Get("/flush", Flush)

	// Wire up named endpoints for each configured proxy
	for _, proxy := range config.Get().Proxies {
		namedAdminGroup := adminGroup.Group(proxy.Name)

		// add cache name middleware to named handler group
		namedAdminGroup.Use(middleware.GenCacheNameMiddleware(proxy.Name))

		for path, handler := range namedEndpoints {
			handlerPath := path
			// if dynamic endpoint configured, add endpoint path parameter
			if proxy.HasEndpointParam {
				handlerPath = "/:e" + handlerPath
			}

			// configure proxy endpoint genHandler
			namedAdminGroup.Get(handlerPath, handler)
		}
	}
}

// namedEndpoints is a map of proxy-specific handler functions and their paths
var namedEndpoints = map[string]fiber.Handler{
	// return stats for a cache by name
	"/stats": Stats,
	// flush the in-memory cache of a proxy by name
	"/flush": Flush,
	// invalidate a given tile without re-priming
	"/invalidate/:z/:x/:y": InvalidateTile,
	// invalidate a given tile and all of its children up to a given max
	// maxZoom defaults to zoom level 12
	"/invalidate/deep/:z/:x/:y":          InvalidateTileDeep,
	"/invalidate/deep/:z/:x/:y/:maxZoom": InvalidateTileDeep,
	// invalidate and prime a given tile
	"/prime/:z/:x/:y": PrimeTile,
	// invalidate and prime a given tile and all of its children up to a given max
	// maxZoom defaults to zoom level 12
	"/prime/deep/:z/:x/:y":          PrimeTileDeep,
	"/prime/deep/:z/:x/:y/:maxZoom": PrimeTileDeep,
}
