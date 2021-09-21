package proxy

import (
	"fmt"
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
		// calculate url and cache key from the configured URL and params
		url, cacheKey, err := replaceParams(ctx, p.TileURL)
		if err != nil {
			ctx.Locals("lod-cache", " :err ")
			util.Error(str.CProxy, str.EBadRequest, err.Error())
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		if cachedTile := cache.Caches.Get(p.Name).Fetch(cacheKey); cachedTile != nil {
			// IF WE HIT A CACHED TILE
			// write the tile to the response body
			_, err := ctx.Write(cachedTile.TileData())
			if err != nil {
				ctx.Locals("lod-cache", "  :err")
				util.Error(str.CProxy, str.EWrite, err.Error(), tileError{
					url:   url,
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
			if err := proxy.Do(ctx, url); err != nil {
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

// replaceParams will substitute URL tile params into the proxy tile URL
func replaceParams(ctx *fiber.Ctx, url string) (string, string, error) {
	z, zErr := ctx.ParamsInt("z")
	if zErr != nil {
		return "", "", zErr
	}

	x, xErr := ctx.ParamsInt("x")
	if xErr != nil {
		return "", "", xErr
	}

	y, yErr := ctx.ParamsInt("y")
	if yErr != nil {
		return "", "", yErr
	}

	replacedUrl := strings.ReplaceAll(url, "{z}", strconv.Itoa(z))
	replacedUrl = strings.ReplaceAll(replacedUrl, "{x}", strconv.Itoa(x))
	replacedUrl = strings.ReplaceAll(replacedUrl, "{y}", strconv.Itoa(y))
	return replacedUrl, fmt.Sprintf("%d/%d/%d", z, x, y), nil
}
