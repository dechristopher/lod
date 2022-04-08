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
	middleware.Wire(r, &p)

	// enable auth middleware if access token configured
	if p.AccessToken != "" {
		proxyGroup.Use(middleware.GenAuthMiddleware(p.AccessToken,
			middleware.Query, false))
	}

	path := handlerEndpointPath
	// if dynamic endpoint configured, add endpoint path parameter
	if p.HasEndpointParam {
		path = "/:e" + path
	}

	// configure proxy endpoint genHandler
	proxyGroup.Get(path, genHandler(p))
}
