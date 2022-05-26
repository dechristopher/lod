package proxy

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/singleflight"

	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/helpers"
	"github.com/dechristopher/lod/packet"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/util"
)

type tileError struct {
	url   string
	proxy config.Proxy
}

var flightGroup singleflight.Group

// genHandler builds a new proxy endpoint handler from configuration
func genHandler(p config.Proxy) fiber.Handler {
	// get cache instance for this proxy
	c := cache.Get(p.Name)

	// handler function to wire to endpoint
	return func(ctx *fiber.Ctx) error {
		return handle(p, c, ctx)
	}
}

// handle proxy requests for the specified proxy config
func handle(p config.Proxy, c *cache.Cache, ctx *fiber.Ctx) error {
	// check presence of configured URL parameters and store
	// their values in a map within the request locals
	helpers.FillParamsMap(p, ctx)

	// build tileUrl and cacheKey from request context and config
	tileUrl, cacheKey, err := buildKeyAndUrl(p, ctx)
	if err != nil {
		// buildKeyAndUrl log their own errors, so no need to here
		return ctx.Status(fiber.StatusBadRequest).SendString("")
	}

	// attempt to fetch the tile from cache before hitting the upstream
	if cachedTile := c.Fetch(cacheKey, ctx); cachedTile != nil {
		// IF WE HIT A CACHED TILE
		if err = returnCachedTile(ctx, p, tileUrl, cachedTile); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString("")
		}
	} else {
		// IF WE MISSED A CACHED TILE
		ctx.Locals(str.LocalCacheStatus, ":miss ")

		// clean up flight group after request is done
		defer flightGroup.Forget(cacheKey)

		// fetch tile via agent proxy, ensuring only a single request is in flight at a given time
		response, errProxy, waited := flightGroup.Do(cacheKey, helpers.FetchUpstream(tileUrl, p))

		if errProxy != nil {
			// return internal server error status if agent proxy request failed in flight
			util.Error(str.CProxy, str.EProxyAgentError, p.Name, cacheKey, errProxy.Error())
			ctx.Locals(str.LocalCacheStatus, ":err-a")
			return ctx.Status(fiber.StatusInternalServerError).SendString("")
		}

		if waited {
			ctx.Locals(str.LocalCacheStatus, ":hit-w")
		}

		// cast interface returned from flight group to a proxyResponse
		proxyResp, ok := response.(helpers.ProxyResponse)

		// sanity check to ensure cast worked properly
		if !ok {
			util.Error(str.CProxy, str.EProxyBadCast, p.Name, cacheKey)
			ctx.Locals(str.LocalCacheStatus, ":err-i")
			return ctx.Status(fiber.StatusInternalServerError).SendString("")
		}

		// write tile data and headers and cache result
		if err = helpers.ProcessResponse(helpers.ProcessResponsePayload{
			Ctx:       ctx,
			Cache:     c,
			Proxy:     p,
			CacheKey:  cacheKey,
			Response:  proxyResp,
			WriteData: true,
		}); err != nil {
			util.Error(str.CProxy, str.EProxyWrite, p.Name, cacheKey, err.Error())
			ctx.Locals(str.LocalCacheStatus, ":err-u")
			// Send internal server error response with empty body if upstream
			// fails to respond or responds with a non-200 status code
			return ctx.Status(fiber.StatusInternalServerError).SendString("")
		}
	}

	return nil
}

// buildKeyAndUrl returns the upstream tile URL and cache key using the given
// proxy configuration and fiber request context
func buildKeyAndUrl(p config.Proxy, ctx *fiber.Ctx) (string, string, error) {
	// calculate url from the configured URL and params
	tileUrl, err := helpers.BuildTileUrl(p, ctx)
	if err != nil {
		ctx.Locals(str.LocalCacheStatus, ":err-t")
		util.Error(str.CProxy, str.ECacheBuildTileUrl, err.Error())
		return "", "", err
	}

	// calculate the cache key for this request using XYZ and URL params
	cacheKey, err := helpers.BuildCacheKey(p, ctx)
	if err != nil {
		ctx.Locals(str.LocalCacheStatus, ":err-c")
		util.Error(str.CProxy, str.ECacheBuildKey, err.Error())
		return "", "", err
	}

	return tileUrl, cacheKey, nil
}

// returnCachedTile is called if the cache contains the requested tile
func returnCachedTile(ctx *fiber.Ctx, p config.Proxy, tileUrl string, cachedTile *packet.TilePacket) error {
	// write the tile to the response body
	_, err := ctx.Write(cachedTile.TileData())
	if err != nil {
		ctx.Locals(str.LocalCacheStatus, ":err-w")
		util.Error(str.CProxy, str.EWrite, err.Error(), tileError{
			url:   tileUrl,
			proxy: p,
		})
		return err
	}

	// set stored headers in response
	for key, val := range cachedTile.Headers() {
		ctx.Set(key, val)
	}

	// remove delete list headers from final response
	p.DoDeleteHeaders(ctx)

	return nil
}
