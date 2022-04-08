package helpers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/tile"
)

// GetTile computes the tile from the request URL
func GetTile(ctx *fiber.Ctx) (*tile.Tile, error) {
	x, xErr := ctx.ParamsInt("x")
	if xErr != nil {
		return nil, xErr
	}

	y, yErr := ctx.ParamsInt("y")
	if yErr != nil {
		return nil, yErr
	}

	zoom, zErr := ctx.ParamsInt("z")
	if zErr != nil {
		return nil, zErr
	}

	return &tile.Tile{
		X:    x,
		Y:    y,
		Zoom: zoom,
	}, nil
}

// BuildCacheKey will put together a cache key from the configured template
func BuildCacheKey(proxy config.Proxy, ctx *fiber.Ctx, t ...tile.Tile) (string, error) {
	var currentTile *tile.Tile
	var err error

	if len(t) == 0 || t == nil {
		currentTile, err = GetTile(ctx)
		if err != nil {
			return "", err
		}
	} else {
		currentTile = &t[0]
	}

	// replace XYZ values in the key template
	key := currentTile.InjectString(proxy.Cache.KeyTemplate)

	// replace dynamic endpoint parameter in cache key if configured
	if proxy.HasEndpointParam && strings.Contains(key, tile.EndpointTemplate) {
		endpoint := ctx.Params("e")
		key = strings.ReplaceAll(key, tile.EndpointTemplate, endpoint)
	}

	// fetch params from context for possible substitution
	paramsMap := GetParamsFromCtx(ctx)
	if paramsMap == nil {
		return key, nil
	}

	// replace params by name in the key template if any exist
	for param, val := range paramsMap {
		key = strings.ReplaceAll(key, fmt.Sprintf("{%s}", param), val)
	}

	return key, nil
}

// FillParamsMap will populate a map local to the request context with configured
// parameter values if any are present in the request
func FillParamsMap(proxy config.Proxy, ctx *fiber.Ctx) {
	paramsMap := make(map[string]string)
	for _, param := range proxy.Params {
		if val := ctx.Query(param.Name, param.Default); val != "" {
			paramsMap[param.Name] = val
		}
	}

	if len(paramsMap) > 0 {
		ctx.Locals("params", paramsMap)
	}
}

// GetParamsFromCtx will attempt to fetch the params map from the request
// context locals if any parameters are present and valid
func GetParamsFromCtx(ctx *fiber.Ctx) map[string]string {
	if ctx.Locals("params") != nil {
		return ctx.Locals("params").(map[string]string)
	}
	return nil
}
