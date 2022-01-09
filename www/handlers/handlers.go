package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/www/handlers/admin"
	"github.com/tile-fund/lod/www/handlers/instance"

	"github.com/tile-fund/lod/www/handlers/proxy"
	"github.com/tile-fund/lod/www/middleware"
)

// Wire builds all the websocket and http routes
// into the fiber app context
func Wire(r *fiber.App) {
	// recover from panics
	r.Use(recover.New())

	// wire instance group handlers
	instance.Wire(r)

	// wire admin group handlers if not disabled
	if !config.Cap.Instance.AdminDisabled {
		admin.Wire(r)
	}

	// wire proxy groups and handlers for each configured proxy
	proxy.Wire(r)

	// Custom 404 page
	middleware.NotFound(r)
}
