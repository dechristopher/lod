package proxy

import (
	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/util"
	"github.com/dechristopher/lod/www/middleware"
)

// Wire proxy group and endpoints for each configured proxy
func Wire(r *fiber.App) {
	for _, p := range config.Get().Proxies {
		wireProxy(r, p)
		util.Info(str.CMain, str.MProxy, p.Cache.MemEnabled, p.Cache.RedisEnabled, p.Name, p.TileURL)
	}
}

const handlerEndpointPath = "/:z/:x/:y.*"

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

	path := handlerEndpointPath
	// if dynamic endpoint configured, add endpoint path parameter
	if p.HasEParam {
		path = "/:e" + path
	}

	// configure CORS preflight genHandler
	proxyGroup.Options(path, preflight)

	// configure proxy endpoint genHandler
	proxyGroup.Get(path, genHandler(p))

	// set common cors headers after handlers to override response from upstream
	proxyGroup.Use(corsHeaders())
}

// corsHeaders sets CORS headers after proxy genHandler execution
func corsHeaders() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Vary(fiber.HeaderOrigin)

		// Set CORS allow methods
		ctx.Set("Access-Control-Allow-Methods", "GET,OPTIONS")
		// Set CORS origin headers
		ctx.Set("Access-Control-Allow-Origin", ctx.Get(fiber.HeaderOrigin))
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
