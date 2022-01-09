package proxy

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
	"github.com/tile-fund/lod/www/middleware"
)

// Wire proxy group and endpoints for each configured proxy
func Wire(r *fiber.App) {
	for _, p := range config.Cap.Proxies {
		wireProxy(r, p)
		util.Info(str.CMain, str.MProxy, p.Name, p.TileURL)
	}
}

// wireProxy configures a new proxy endpoint from the configuration under
// a named Router group
func wireProxy(r *fiber.App, p config.Proxy) {
	// genHandler group for this proxy instance
	proxyGroup := r.Group(p.Name)

	// wire middleware for proxy group
	middleware.Wire(r, p)

	// enable auth middleware if admin token configured
	if p.AccessToken != "" {
		proxyGroup.Use(middleware.GenAuthMiddleware(p.AccessToken,
			middleware.Query, true))
	}

	// configure CORS preflight genHandler
	proxyGroup.Options("/:z/:x/:y.*", preflight)

	// configure proxy endpoint genHandler
	proxyGroup.Get("/:z/:x/:y.*", genHandler(p))

	// set common cors headers after handlers to override response from upstream
	proxyGroup.Use(corsHeaders(p))
}

// corsHeaders sets cord headers after proxy genHandler execution
func corsHeaders(p config.Proxy) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Set CORS allow methods
		ctx.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		// Set CORS origin headers
		ctx.Set("Access-Control-Allow-Origin", config.CorsOrigins(p))
		return nil
	}
}

// preflight genHandler for CORS OPTIONS requests
func preflight(ctx *fiber.Ctx) error {
	// Tell client that this pre-flight info is valid for 20 days
	ctx.Set("Access-Control-Max-Age", "1728000")
	ctx.Set("Content-Type", "text/plain charset=UTF-8")
	ctx.Set("Content-Length", "0")
	return ctx.SendStatus(fiber.StatusNoContent)
}
