package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/env"
	"github.com/dechristopher/lod/str"
)

// Wire attaches all middleware to the given router
func Wire(r fiber.Router, proxy *config.Proxy) {
	r.Use(requestid.New())

	// Compress responses for non-tiles, use tileserver compression and encoding
	if proxy == nil {
		r.Use(compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}))
	}

	// Configure CORS
	var origins string

	if proxy != nil {
		origins = proxy.CorsOrigins
	}

	if origins == "" {
		origins = "*"
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodHead,
			fiber.MethodOptions,
		}, ","),
	}))
}

// AuthType used for auth middleware generation
type AuthType string

const (
	// Bearer token header auth
	Bearer AuthType = "bearer"
	// Query string in URL (?token=)
	Query AuthType = "query"
)

// GenCacheNameMiddleware builds a middleware that adds the proxy name to the
// request context locals so that named admin endpoints can look up the proxy
// they handle requests for
func GenCacheNameMiddleware(cacheName string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Locals(str.LocalCacheName, cacheName)
		return ctx.Next()
	}
}

// GenAuthMiddleware builds a middleware that checks for valid tokens
func GenAuthMiddleware(token string, authType AuthType, notFound bool) fiber.Handler {
	bearer := fmt.Sprintf("Bearer %s", token)

	var authCheck func(ctx *fiber.Ctx, token string) bool

	if authType == Bearer {
		authCheck = func(ctx *fiber.Ctx, token string) bool {
			return ctx.GetReqHeaders()[fiber.HeaderAuthorization] == bearer
		}
	} else {
		authCheck = func(ctx *fiber.Ctx, token string) bool {
			return ctx.Query("token") == token
		}
	}

	return func(ctx *fiber.Ctx) error {
		if authCheck(ctx, token) {
			// continue normally if checks succeed
			return ctx.Next()
		}

		ctx.Locals(str.LocalCacheStatus, ":nauth")

		if !env.IsProd() {
			// provide useful error messages when running in dev mode
			return ctx.Status(fiber.StatusUnauthorized).JSON(map[string]string{
				"status":  "error",
				"message": "failed to auth, invalid token supplied",
			})
		}

		if notFound {
			// otherwise, pretend nothing exists if notFound is set
			return ctx.Status(fiber.StatusNotFound).SendString("")
		}

		// return empty 401 if notFound is not set
		return ctx.Status(fiber.StatusUnauthorized).SendString("")
	}
}

// NotFound wires the final 404 handler after all other
// handlers are defined. Acts as the final fallback.
func NotFound(r *fiber.App) {
	r.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString("")
	})
}
