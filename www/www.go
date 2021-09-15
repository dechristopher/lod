package www

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/tile-fund/lod/www/proxy"

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
		CaseSensitive:         true,
		DisableStartupMessage: true,
		ProxyHeader:           "X-Forwarded-For",
		ReadTimeout:           time.Second * 10,
		WriteTimeout:          time.Second * 30,
		IdleTimeout:           time.Hour,
	})

	// wire up all route handlers
	err := wireHandlers(r)
	if err != nil {
		util.Error(str.CMain, str.EWire, err.Error())
	}

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
func wireHandlers(r *fiber.App) error {
	// recover from panics
	r.Use(recover.New())

	lodGroup := r.Group("/lod")

	// wire up all middleware components
	middleware.Wire(lodGroup)

	// capabilities endpoint shows configuration summary
	lodGroup.Get("/capabilities", handlers.Capabilities)

	// JSON service health / status handler
	lodGroup.Get("/status", handlers.Status)

	// configure proxy endpoints for each configured proxy
	for _, p := range config.Cap.Proxies {
		err := proxy.WireProxy(r, p)
		if err != nil {
			return err
		}
		util.Info(str.CMain, str.MProxy, p.Name, p.TileURL)
	}

	// Custom 404 page
	middleware.NotFound(r)

	return nil
}
