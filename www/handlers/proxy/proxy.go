package proxy

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/tile-fund/lod/cache"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/helpers"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
)

type tileError struct {
	url   string
	proxy config.Proxy
}

// genHandler builds a new proxy endpoint handler from configuration
func genHandler(p config.Proxy) fiber.Handler {
	// preconfigure cache on boot
	cache.Get(p.Name)

	// handler function to wire to endpoint
	return func(ctx *fiber.Ctx) error {
		return handle(p, ctx)
	}
}

// handle proxy requests for the specified proxy config
func handle(p config.Proxy, ctx *fiber.Ctx) error {
	// check presence of configured URL parameters and store
	// their values in a map within the request locals
	helpers.FillParamsMap(p, ctx)

	// calculate url from the configured URL and params
	tileUrl, err := buildTileUrl(p, ctx)
	if err != nil {
		ctx.Locals("lod-cache", " :err ")
		util.Error(str.CProxy, str.EBadRequest, err.Error())
		return ctx.Status(fiber.StatusBadRequest).SendString("")
	}

	// calculate the cache key for this request using XYZ and URL params
	cacheKey, err := helpers.BuildCacheKey(p, ctx)
	if err != nil {
		ctx.Locals("lod-cache", "  :err")
		util.Error(str.CProxy, str.ECacheBuildKey, err.Error())
		return ctx.Status(fiber.StatusInternalServerError).SendString("")
	}

	if cachedTile := cache.Get(p.Name).Fetch(cacheKey); cachedTile != nil {
		// IF WE HIT A CACHED TILE
		// write the tile to the response body
		_, err := ctx.Write(cachedTile.TileData())
		if err != nil {
			ctx.Locals("lod-cache", "  :err")
			util.Error(str.CProxy, str.EWrite, err.Error(), tileError{
				url:   tileUrl,
				proxy: p,
			})
			return ctx.Status(fiber.StatusInternalServerError).SendString("")
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
			go cache.Get(p.Name).EncodeSet(cacheKey, tileData, headers)
		}
	}

	// Remove server header from response
	ctx.Response().Header.Del(fiber.HeaderServer)

	return ctx.Next()
}

// buildTileUrl will substitute URL tile params into the proxy tile URL
func buildTileUrl(proxy config.Proxy, ctx *fiber.Ctx) (string, error) {
	currentTile, err := helpers.GetTile(ctx)
	if err != nil {
		return "", err
	}

	// replace XYZ values in the tile URL
	baseUrl := currentTile.InjectString(proxy.TileURL)

	// fetch params from context for possible addition to URL
	paramsMap := helpers.GetParamsFromCtx(ctx)

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
