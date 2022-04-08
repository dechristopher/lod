package admin

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/helpers"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/util"
)

type invalidateAndPrimePayload struct {
	MaxZoom      int    // max zoom to deepen to
	Prime        bool   // whether to re-prime the tiles after invalidating
	ErrorMessage string // error message template string
	InfoMessage  string // info message template string
}

// InvalidateAndPrime is the parent function to all invalidation and priming endpoints
func InvalidateAndPrime(ctx *fiber.Ctx, payload invalidateAndPrimePayload) error {
	// get cache by name for this request if one is configured
	c := cache.Get(ctx.Params("name"))
	if c == nil {
		util.Error(str.CAdmin, payload.ErrorMessage, "unknown", "invalid proxy name")
		return ctx.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"status": "failed",
			"error":  "invalid proxy name provided",
		})
	}

	// fill params map to augment param segmentation behavior present in proxy endpoint
	helpers.FillParamsMap(*c.Proxy, ctx)

	// get requested tile from context
	tile, err := helpers.GetTile(ctx)
	if err != nil {
		util.Error(str.CAdmin, payload.ErrorMessage, "unknown", err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"status":  "failed",
			"error":   "invalid tile requested",
			"message": err.Error(),
		})
	}

	// determine max zoom based on parameter
	maxZoom, err := ctx.ParamsInt("maxZoom", payload.MaxZoom)
	if err != nil {
		util.Error(str.CAdmin, payload.ErrorMessage, tile.String(), err.Error())
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"status":  "failed",
			"error":   "internal server error",
			"message": err.Error(),
		})
	}

	// calculate all necessary tiles for this operation
	tiles := tile.DeepChildren(maxZoom)

	failed := 0

	if !payload.Prime {
		// simply invalidate en masse
		for _, t := range tiles {
			key, err := helpers.BuildCacheKey(*c.Proxy, ctx, t)
			if err != nil {
				util.Debug(str.CAdmin, str.DInvalidateFail, tile.String(), err.Error())
				failed++
				continue
			}
			errInv := c.Invalidate(key, ctx.Context())
			if errInv != nil {
				util.Debug(str.CAdmin, str.DInvalidateFail, tile.String(), errInv)
				failed++
			}
		}
	}
	//else {
	// fetch and prime in place for the given tile to avoid invalidating tiles
	// en masse and having missing tiles in the cache during the prime
	//}

	util.Info(str.CAdmin, payload.InfoMessage, tile.String(), maxZoom, len(tiles))
	return ctx.JSON(map[string]string{
		"status":    "ok",
		"completed": fmt.Sprintf("%d/%d tiles", len(tiles)-failed, len(tiles)),
	})
}

// InvalidateTile will invalidate a tile from the caches if it exists
func InvalidateTile(ctx *fiber.Ctx) error {
	return InvalidateAndPrime(ctx, invalidateAndPrimePayload{
		MaxZoom:      0, // only this one tile
		Prime:        false,
		ErrorMessage: str.EInvalidateTile,
		InfoMessage:  str.MInvalidateTile,
	})
}

// InvalidateTileDeep performs the same action as PrimeTile, only on all child tiles
// resultant of the provided tile using iterative deepening up to a given max zoom
func InvalidateTileDeep(ctx *fiber.Ctx) error {
	return InvalidateAndPrime(ctx, invalidateAndPrimePayload{
		MaxZoom:      12, // default max of 12
		Prime:        false,
		ErrorMessage: str.EInvalidateTileDeep,
		InfoMessage:  str.MInvalidateTileDeep,
	})
}

// PrimeTile will invalidate a tile from the caches if it exists and will make
// a request to re-prime the tile fresh from the upstream tileserver
func PrimeTile(ctx *fiber.Ctx) error {
	return InvalidateAndPrime(ctx, invalidateAndPrimePayload{
		MaxZoom:      0, // only this one tile
		Prime:        true,
		ErrorMessage: str.EPrimeTile,
		InfoMessage:  str.MPrimeTile,
	})
}

// PrimeTileDeep performs the same action as PrimeTile, only on all child tiles
// resultant of the provided tile using iterative deepening up to a given max zoom
func PrimeTileDeep(ctx *fiber.Ctx) error {
	return InvalidateAndPrime(ctx, invalidateAndPrimePayload{
		MaxZoom:      12, // default max of 12
		Prime:        true,
		ErrorMessage: str.EPrimeTileDeep,
		InfoMessage:  str.MPrimeTileDeep,
	})
}
