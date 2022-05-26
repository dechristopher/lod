package admin

import (
	"sync"

	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/helpers"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/tile"
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
	c := cache.Get(ctx.Locals(str.LocalCacheName).(string))
	if c == nil {
		util.Error(str.CAdmin, payload.ErrorMessage, "unknown", "invalid proxy name")
		return ctx.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"status": "failed",
			"error":  "invalid proxy name provided",
		})
	}

	// fill params map to augment param segmentation behavior present in proxy endpoint
	helpers.FillParamsMap(*c.Proxy, ctx)

	// get requested reqTile from context
	reqTile, err := helpers.GetTile(ctx)
	if err != nil {
		util.Error(str.CAdmin, payload.ErrorMessage, "unknown", err.Error())
		return ctx.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"status":  "failed",
			"error":   "invalid reqTile requested",
			"message": err.Error(),
		})
	}

	// determine max zoom based on parameter
	maxZoom, err := ctx.ParamsInt("maxZoom", payload.MaxZoom)
	if err != nil {
		util.Error(str.CAdmin, payload.ErrorMessage, reqTile.String(), err.Error())
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"status":  "failed",
			"error":   "internal server error",
			"message": err.Error(),
		})
	}

	// calculate all necessary tiles for this operation
	tiles := reqTile.DeepChildren(maxZoom)

	util.Debug(str.CAdmin, str.DCalcTiles, c.Proxy.Name,
		len(tiles), reqTile.String(), maxZoom)

	succeeded := 0

	if !payload.Prime {
		// simply invalidate en masse
		for _, t := range tiles {
			key, err := helpers.BuildCacheKey(*c.Proxy, ctx, t)
			if err != nil {
				util.Debug(str.CAdmin, str.DInvalidateFail, reqTile.String(), err.Error())
				continue
			}
			errInv := c.Invalidate(key, ctx.Context())
			if errInv != nil {
				util.Debug(str.CAdmin, str.DInvalidateFail, reqTile.String(), errInv)
			}
			succeeded++
		}
	} else {
		// fetch and prime in place for the given tile to avoid invalidating tiles
		// en masse and having missing tiles in the cache during the priming period
		wg := &sync.WaitGroup{}
		wg.Add(c.Proxy.NumWorkers)

		jobs := make(chan tile.Tile, len(tiles))
		successes := make(chan bool, len(tiles))

		// spin up workers to make agent-proxied requests to the upstream
		for numWorkers := 0; numWorkers < c.Proxy.NumWorkers; numWorkers++ {
			go tileWorker(tileWorkerPayload{
				jobs:      jobs,
				successes: successes,
				cache:     c,
				ctx:       ctx,
				waitGroup: wg,
			})
		}

		// submit jobs to workers
		for _, tileJob := range tiles {
			jobs <- tileJob
		}

		// signal that we're out of tiles to prime
		close(jobs)

		// wait until workers finish
		wg.Wait()

		// close successes channel after workers finish
		close(successes)

		// count successfully primed tiles
		for range successes {
			succeeded++
		}
	}

	status := "ok"
	if succeeded != len(tiles) {
		status = "failed"
	}

	util.Info(str.CAdmin, payload.InfoMessage, reqTile.String(), maxZoom, len(tiles))
	return ctx.JSON(map[string]interface{}{
		"attempted": len(tiles),
		"primed":    succeeded,
		"status":    status,
	})
}

// tileWorkerPayload is a struct containing all the ingredients
// needed for a tileWorker to operate on its job queue
type tileWorkerPayload struct {
	jobs      <-chan tile.Tile
	successes chan<- bool
	cache     *cache.Cache
	ctx       *fiber.Ctx
	waitGroup *sync.WaitGroup
}

// tileWorker is a worker function that's spun up during requests to prime and
// invalidate batches of tiles
func tileWorker(payload tileWorkerPayload) {
	defer payload.waitGroup.Done()

	for tileJob := range payload.jobs {
		url, err := helpers.BuildTileUrl(*payload.cache.Proxy, payload.ctx, tileJob)
		if err != nil {
			util.Debug(str.CAdmin, str.DPrimeFail, tileJob.String(), err.Error())
			continue
		}

		cacheKey, err := helpers.BuildCacheKey(*payload.cache.Proxy, payload.ctx, tileJob)
		if err != nil {
			util.Debug(str.CAdmin, str.DPrimeFail, tileJob.String(), err.Error())
			continue
		}

		response, errProxy := helpers.FetchUpstream(url, *payload.cache.Proxy)()
		if errProxy != nil {
			util.Debug(str.CAdmin, str.DPrimeFail, tileJob.String(), err.Error())
			continue
		}

		// cast interface returned from flight group to a proxyResponse
		proxyResp, ok := response.(helpers.ProxyResponse)

		// sanity check to ensure cast worked properly
		if !ok {
			util.Debug(str.CAdmin, str.DPrimeFail, tileJob.String(), err.Error())
			continue
		}

		// write reqTile data and headers and cache result
		if err = helpers.ProcessResponse(helpers.ProcessResponsePayload{
			Ctx:       payload.ctx,
			Cache:     payload.cache,
			Proxy:     *payload.cache.Proxy,
			CacheKey:  cacheKey,
			Response:  proxyResp,
			WriteData: true,
		}); err == nil {
			// signal successful tile
			payload.successes <- true
		}
	}
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
