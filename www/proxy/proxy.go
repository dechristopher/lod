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

// WireProxy configures a new proxy endpoint from the configuration under
// a named Router group
func WireProxy(r *fiber.App, p config.Proxy) error {

	proxyGroup := r.Group(p.Name)

	// wire middleware for proxy group
	middleware.Wire(r, p)

	// configure CORS preflight handler
	proxyGroup.Options("/:z/:x/:y.pbf", preflight)

	// configure proxy endpoint handler
	proxyGroup.Get("/:z/:x/:y.pbf", handler(p))

	// set common cors headers after handlers to override response from upstream
	proxyGroup.Use(corsHeaders(p))

	return nil
}

// corsHeaders sets cord headers after proxy handler execution
func corsHeaders(p config.Proxy) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set CORS allow methods
		c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		// Set CORS origin headers
		c.Set("Access-Control-Allow-Origin", config.CorsOrigins(p))
		return nil
	}
}

// preflight handler for CORS OPTIONS requests
func preflight(c *fiber.Ctx) error {
	// Tell client that this pre-flight info is valid for 20 days
	c.Set("Access-Control-Max-Age", "1728000")
	c.Set("Content-Type", "text/plain charset=UTF-8")
	c.Set("Content-Length", "0")
	return c.SendStatus(fiber.StatusNoContent)
}

// Build a new proxy endpoint handler from configuration
func handler(p config.Proxy) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// calculate url and cache key from the configured URL and params
		url, cacheKey, err := replaceParams(c, p.TileURL)
		if err != nil {
			c.Locals("lod-cache", " :err ")
			util.Error(str.CProxy, str.EBadRequest, err.Error())
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if cachedTile := cache.Caches.Get(p.Name).Fetch(cacheKey); cachedTile != nil {
			// IF WE HIT A CACHED TILE
			// write the tile to the response body
			_, err := c.Write(cachedTile.Data)
			if err != nil {
				c.Locals("lod-cache", "  :err")
				util.Error(str.CProxy, str.EWrite, err.Error(), tileError{
					url:   url,
					proxy: p,
				})
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			c.Locals("lod-cache", " :hit ")

			// set stored headers in response
			for key, val := range cachedTile.Headers {
				c.Set(key, val)
			}
		} else {
			// IF WE MISSED A CACHED TILE
			c.Locals("lod-cache", " :miss")
			// perform request to tile URL
			if err := proxy.Do(c, url); err != nil {
				return err
			}

			// spin off a routine to cache the tile without blocking the response
			if len(c.Response().Body()) > 0 {
				// copy tile data into separate slice, so we don't lose the reference
				tileData := make([]byte, len(c.Response().Body()))
				copy(tileData, c.Response().Body())

				tile := cache.Tile{
					Data:    tileData,
					Headers: map[string]string{},
				}

				// Store configured headers into the tile cache for this tile
				p.PopulateHeaders(c, tile.Headers)

				go cache.Caches.Get(p.Name).Set(cacheKey, tile)
			}
		}

		// Remove server header from response
		c.Response().Header.Del(fiber.HeaderServer)

		return c.Next()
	}
}

// replaceParams will substitute URL tile params into the proxy tile URL
func replaceParams(c *fiber.Ctx, url string) (string, string, error) {
	z, zErr := c.ParamsInt("z")
	if zErr != nil {
		return "", "", zErr
	}

	x, xErr := c.ParamsInt("x")
	if xErr != nil {
		return "", "", xErr
	}

	y, yErr := c.ParamsInt("y")
	if yErr != nil {
		return "", "", yErr
	}

	replacedUrl := strings.ReplaceAll(url, "{z}", strconv.Itoa(z))
	replacedUrl = strings.ReplaceAll(replacedUrl, "{x}", strconv.Itoa(x))
	replacedUrl = strings.ReplaceAll(replacedUrl, "{y}", strconv.Itoa(y))
	return replacedUrl, fmt.Sprintf("%d/%d/%d", z, x, y), nil
}
