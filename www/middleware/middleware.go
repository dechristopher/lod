package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/tile-fund/lod/config"

	"github.com/tile-fund/lod/env"
)

const logFormatProd = "${ip} ${header:x-forwarded-for} ${header:x-real-ip} " +
	"[${time}] ${pid} ${locals:requestid}${locals:lod-cache} \"${method} ${path} ${protocol}\" " +
	"${status} ${latency} ${bytesSent}b \"${referrer}\" \"${ua}\"\n"

const logFormatDev = "${ip} [${time}]${locals:lod-cache} \"${method} ${path} ${protocol}\" " +
	"${status} ${latency} ${bytesSent}b\n"

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

	// STDOUT request logger
	r.Use(logger.New(logger.Config{
		TimeZone:   "local",
		TimeFormat: "2006-01-02T15:04:05-0700",
		Format:     logFormat(),
		Output:     os.Stdout,
	}))
}

// NotFound wires the final 404 handler after all other
// handlers are defined. Acts as the final fallback.
func NotFound(r *fiber.App) {
	r.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})
}

// logFormat returns the HTTP log format for the
// configured fiber logger middleware
func logFormat() string {
	if env.IsProd() {
		return logFormatProd
	}
	return logFormatDev
}
