package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/www/handlers/admin"
	"github.com/dechristopher/lod/www/handlers/proxy"
	"github.com/dechristopher/lod/www/middleware"
)

// Wire builds all the websocket and http routes
// into the fiber app context
func Wire(r *fiber.App) {
	// recover from panics
	r.Use(recover.New())

	// wire admin group handlers if not disabled
	if !config.Get().Instance.AdminDisabled {
		admin.Wire(r)
	}

	// wire proxy groups and handlers for each configured proxy
	proxy.Wire(r)

	// Custom 404 page
	middleware.NotFound(r)
}
