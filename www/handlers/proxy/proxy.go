package proxy

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/tile-fund/lod/cache"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
	"github.com/tile-fund/lod/www/middleware"
)

type tileError struct {
	url   string
	proxy config.Proxy
}

// wireProxy configures a new proxy endpoint from the configuration under
// a named Router group
func wireProxy(r *fiber.App, p config.Proxy) {
	// handler group for this proxy instance
	proxyGroup := r.Group(p.Name)

	// wire middleware for proxy group
	middleware.Wire(r, p)

	// configure CORS preflight handler
	proxyGroup.Options("/:z/:x/:y.pbf", preflight)

	// configure proxy endpoint handler
	proxyGroup.Get("/:z/:x/:y.pbf", handler(p))

	// set common cors headers after handlers to override response from upstream
	proxyGroup.Use(corsHeaders(p))
}

// corsHeaders sets cord headers after proxy handler execution
func corsHeaders(p config.Proxy) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Set CORS allow methods
		ctx.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		// Set CORS origin headers
		ctx.Set("Access-Control-Allow-Origin", config.CorsOrigins(p))
		return nil
	}
}

// preflight handler for CORS OPTIONS requests
func preflight(ctx *fiber.Ctx) error {
	// Tell client that this pre-flight info is valid for 20 days
	ctx.Set("Access-Control-Max-Age", "1728000")
	ctx.Set("Content-Type", "text/plain charset=UTF-8")
	ctx.Set("Content-Length", "0")
	return ctx.SendStatus(fiber.StatusNoContent)
}

// Build a new proxy endpoint handler from configuration
func handler(p config.Proxy) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// check presence of configured URL parameters and store
		// their values in a map within the request locals
		fillParamsMap(p, ctx)

		// calculate url from the configured URL and params
		tileUrl, err := getBaseUrl(p, ctx)
		if err != nil {
			ctx.Locals("lod-cache", " :err ")
			util.Error(str.CProxy, str.EBadRequest, err.Error())
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		// calculate the cache key for this request using XYZ and URL params
		cacheKey, err := buildCacheKey(p, ctx)

		if cachedTile := cache.Caches.Get(p.Name).Fetch(cacheKey); cachedTile != nil {
			// IF WE HIT A CACHED TILE
			// write the tile to the response body
			_, err := ctx.Write(cachedTile.TileData())
			if err != nil {
				ctx.Locals("lod-cache", "  :err")
				util.Error(str.CProxy, str.EWrite, err.Error(), tileError{
					url:   tileUrl,
					proxy: p,
				})
				return ctx.SendStatus(fiber.StatusInternalServerError)
			}

			ctx.Locals("lod-cache", " :hit ")

			// set stored headers in response
			for key, val := range cachedTile.Headers() {
				ctx.Set(key, val)
			}
		} else {
			// IF WE MISSED A CACHED TILE
			ctx.Locals("lod-cache", " :miss")
			// perform request to tile URL
			if err := proxy.Do(ctx, tileUrl); err != nil {
				return err
			}

			if len(ctx.Response().Body()) > 0 {
				// copy tile data into separate slice, so we don't lose the reference
				tileData := make([]byte, len(ctx.Response().Body()))
				copy(tileData, ctx.Response().Body())

				headers := map[string]string{}
				// Store configured headers into the tile cache for this tile
				p.PopulateHeaders(ctx, headers)

				// Delete headers from the final response that are on the DelHeaders list
				// if we got them from the tileserver. This can be used to prevent leaking
				// internals of the tileserver if you don't control what it returns
				p.DeleteHeaders(ctx)

				// spin off a routine to cache the tile without blocking the response
				go cache.Caches.Get(p.Name).EncodeSet(cacheKey, tileData, headers)
			}
		}

		// Remove server header from response
		ctx.Response().Header.Del(fiber.HeaderServer)

		return ctx.Next()
	}
}

// fillParamsMap will populate a map local to the request context with configured
// parameter values if any are present in the request
func fillParamsMap(proxy config.Proxy, ctx *fiber.Ctx) {
	paramsMap := make(map[string]string)
	for _, param := range proxy.Params {
		if val := ctx.Params(param.Name, param.Default); val != "" {
			paramsMap[param.Name] = val
		}
	}

	if len(paramsMap) > 0 {
		ctx.Locals("params", paramsMap)
	}
}

// getBaseUrl will substitute URL tile params into the proxy tile URL
func getBaseUrl(proxy config.Proxy, ctx *fiber.Ctx) (string, error) {
	x, y, z, err := getXYZ(ctx)
	if err != nil {
		return "", err
	}

	// replace XYZ values in the tile URL
	baseUrl := replaceXYZ(proxy.TileURL, x, y, z)

	// fetch params from context for possible addition to URL
	paramsMap := getParamsFromCtx(ctx)

	// if no query parameters, return baseUrl
	if paramsMap == nil {
		return baseUrl, nil
	}

	// parse baseURL to add URL parameters
	paramUrl, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	// replace params by name in the key template if any exist
	for param, val := range paramsMap {
		params.Add(param, val)
	}

	// set encoded params in URL
	paramUrl.RawQuery = params.Encode()

	// return generated URL with substitutions for query parameters
	return paramUrl.String(), nil
}

// buildCacheKey will put together a cache key from the configured template
func buildCacheKey(proxy config.Proxy, ctx *fiber.Ctx) (string, error) {
	x, y, z, err := getXYZ(ctx)
	if err != nil {
		return "", err
	}

	// replace XYZ values in the key template
	key := replaceXYZ(proxy.Cache.KeyTemplate, x, y, z)

	// fetch params from context for possible substitution
	paramsMap := getParamsFromCtx(ctx)
	if paramsMap == nil {
		return key, nil
	}

	// replace params by name in the key template if any exist
	for param, val := range paramsMap {
		key = strings.ReplaceAll(key, fmt.Sprintf("{%s}", param), val)
	}

	return key, nil
}

// getXYZ returns the XYZ values as integers from the request URL
func getXYZ(ctx *fiber.Ctx) (int, int, int, error) {
	x, xErr := ctx.ParamsInt("x")
	if xErr != nil {
		return 0, 0, 0, xErr
	}

	y, yErr := ctx.ParamsInt("y")
	if yErr != nil {
		return 0, 0, 0, yErr
	}

	z, zErr := ctx.ParamsInt("z")
	if zErr != nil {
		return 0, 0, 0, zErr
	}

	return x, y, z, nil
}

// getParamsFromCtx will attempt to fetch the params map from the request
// context locals if any parameters are present and valid
func getParamsFromCtx(ctx *fiber.Ctx) map[string]string {
	if ctx.Locals("params") != nil {
		return ctx.Locals("params").(map[string]string)
	}
	return nil
}

// replaceXYZ fills the {x}, {y}, and {z} tokens in a template URL
//  or cache key with the provided values
func replaceXYZ(base string, x, y, z int) string {
	base = strings.ReplaceAll(base, "{x}", strconv.Itoa(x))
	base = strings.ReplaceAll(base, "{y}", strconv.Itoa(y))
	return strings.ReplaceAll(base, "{z}", strconv.Itoa(z))
}
