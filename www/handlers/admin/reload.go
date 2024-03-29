package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/util"
)

// ReloadCapabilities performs a config reload, picking up any
// changes to the instance capabilities configuration.
func ReloadCapabilities(ctx *fiber.Ctx) error {
	// reload config and update instance capabilities
	err := config.Load()
	if err != nil {
		return errorReload(ctx, err)
	}

	// reinitialize cache instances
	err = cache.Init()
	if err != nil {
		return errorReload(ctx, err)
	}

	util.Info(str.CAdmin, str.MReload)
	return ctx.JSON(map[string]string{
		"status": "ok",
		"file":   *config.File,
	})
}

func errorReload(ctx *fiber.Ctx, err error) error {
	util.Error(str.CAdmin, str.EReload, err.Error())
	return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{
		"status": "failed",
		"file":   *config.File,
		"error":  err.Error(),
	})
}
