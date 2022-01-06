package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/cache"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
)

// Flush an entire proxy cache by name, or all caches by providing a name of "all"
func Flush(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	if name == "" {
		// quit early if no name provided
		ctx.Status(400)
		return ctx.JSON(map[string]string{
			"status": "bad request",
		})
	}

	// search for proxy with given name
	for _, proxy := range config.Cap.Proxies {
		if proxy.Name == name || name == "all" {
			// flush the proxy's internal cache
			err := cache.Get(proxy.Name).Flush()

			if err != nil {
				util.Error(str.CAdmin, str.ECacheFlush, name, err.Error())
				ctx.Status(500)
				return ctx.JSON(map[string]string{
					"status": "failed",
					"error":  err.Error(),
				})
			}

			if name != "all" {
				ctx.Status(200)
				return ctx.JSON(map[string]string{
					"status": "ok",
				})
			}
		}
	}

	if name == "all" {
		ctx.Status(200)
		return ctx.JSON(map[string]string{
			"status": "ok",
		})
	}

	// 404 if no proxy found with given name
	ctx.Status(404)
	return ctx.JSON(map[string]string{
		"status": "no cache with given name",
	})
}
