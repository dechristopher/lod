package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/util"
)

// Flush an entire proxy cache by name, or all caches
func Flush(ctx *fiber.Ctx) error {
	if ctx.Path() == "/admin/flush" {
		// flush all configured proxies
		for _, proxy := range config.Get().Proxies {
			err := cache.Get(proxy.Name).FlushInternal()

			if err != nil {
				util.Error(str.CAdmin, str.ECacheFlush, proxy.Name, err.Error())
				return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
					"status": "failed",
					"error":  err.Error(),
				})
			}
		}

		return ctx.JSON(map[string]string{
			"status": "ok",
		})
	}

	name := ctx.Params("name")
	if name == "" {
		// quit early if no name provided
		return ctx.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"status": "bad request, no proxy name provided",
		})
	}

	// search for proxy with given name
	for _, proxy := range config.Get().Proxies {
		if proxy.Name == name {
			// flush the proxy's internal cache
			err := cache.Get(proxy.Name).FlushInternal()

			if err != nil {
				util.Error(str.CAdmin, str.ECacheFlush, name, err.Error())
				return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
					"status": "failed",
					"error":  err.Error(),
				})
			}

			return ctx.JSON(map[string]string{
				"status": "ok",
			})
		}
	}

	// 404 if no proxy found with given name
	return ctx.Status(fiber.StatusNotFound).JSON(map[string]string{
		"status": "no proxy configured with given name",
	})
}
