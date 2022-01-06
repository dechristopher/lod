package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/tile-fund/lod/config"
)

// Wire attaches all middleware to the given router
func Wire(r fiber.Router, proxy ...config.Proxy) {
	r.Use(requestid.New())

	// Compress responses for non-tiles, use tileserver compression and encoding
	if len(proxy) == 0 {
		r.Use(compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}))
	}

	// Configure CORS for non-tiles
	if len(proxy) == 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins: "*",
			AllowHeaders: "Origin, Content-Type, Accept",
		}))
	}
}

// NotFound wires the final 404 handler after all other
// handlers are defined. Acts as the final fallback.
func NotFound(r *fiber.App) {
	r.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})
}
