package www

import (
	"log"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/tile-fund/lod/config"

	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
	"github.com/tile-fund/lod/www/handlers"
	"github.com/tile-fund/lod/www/middleware"
)

// Serve all public endpoints
func Serve() {
	r := fiber.New(fiber.Config{
		ServerHeader:          "LODv" + config.Version,
		CaseSensitive:         true,
		DisableStartupMessage: true,
	})

	// wire up all route handlers
	wireHandlers(r)

	// Graceful shutdown with SIGINT
	// SIGTERM and others will hard kill
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		util.Info(str.CMain, str.MShutdown)
		_ = r.Shutdown()
	}()

	util.Info(str.CMain, str.MStarted, util.TimeSinceBoot(),
		env.GetEnv(), config.GetPort())

	// listen for connections on primary listening port
	if err := r.Listen(config.GetListenPort()); err != nil {
		log.Println(err)
	}

	// Exit cleanly
	util.Info(str.CMain, str.MExit)
	os.Exit(0)
}

// wireHandlers builds all the websocket and http routes
// into the fiber app context
func wireHandlers(r *fiber.App) {
	// recover from panics
	r.Use(recover.New())

	// wire up all middleware components
	middleware.Wire(r)

	// capabilities endpoint shows configuration summary
	// r.Get("/capabilities", TODO)

	// JSON service health / status handler
	r.Get("/status", handlers.StatusHandler)

	// Custom 404 page
	middleware.NotFound(r)
}
