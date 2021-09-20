package proxy

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
)

// Wire proxy group and endpoints for each configured proxy
func Wire(r *fiber.App) {
	for _, p := range config.Cap.Proxies {
		wireProxy(r, p)
		util.Info(str.CMain, str.MProxy, p.Name, p.TileURL)
	}
}
