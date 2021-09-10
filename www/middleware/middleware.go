package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/util"
)

const logFormatProd = "${ip} ${header:x-forwarded-for} ${header:x-real-ip} " +
	"[${time}] ${pid} ${locals:requestid} \"${method} ${path} ${protocol}\" " +
	"${status} ${latency} \"${referrer}\" \"${ua}\"\n"

const logFormatDev = "${ip} [${time}] \"${method} ${path} ${protocol}\" " +
	"${status} ${latency}\n"

// Wire attaches all middleware to the given router
func Wire(r fiber.Router) {
	r.Use(requestid.New())

	// Compress responses
	r.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: util.CorsOrigins(),
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// STDOUT request logger
	r.Use(logger.New(logger.Config{
		// For more options, see the Config section
		TimeZone:   "local",
		TimeFormat: "2006-01-02T15:04:05-0700",
		Format:     logFormat(),
		Output:     os.Stdout,
	}))

	// set browser id cookie
	r.Use(func(c *fiber.Ctx) error {
		if c.Cookies("bid") == "" {
			c.Cookie(&fiber.Cookie{
				Name:     "bid",
				Value:    util.GenerateCode(16),
				Path:     "/",
				Domain:   "",
				MaxAge:   0,
				Secure:   !env.IsDev(),
				HTTPOnly: true,
				SameSite: "Strict",
			})
		}
		return c.Next()
	})
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
