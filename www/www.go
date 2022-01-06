package www

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/tile-fund/lod/config"
	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
	"github.com/tile-fund/lod/www/handlers"
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
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			util.Error(str.CMain, str.ERequest, ctx.String(), err.Error())

			// send JSON error output if in dev mode
			if env.IsDev() {
				ctx.Status(500)
				return ctx.JSON(map[string]string{
					"status": "internal error",
					"error":  err.Error(),
				})
			}

			// otherwise, simply return 500
			return ctx.SendStatus(500)
		},
	})

	// STDOUT request logger
	r.Use(logger.New(logger.Config{
		TimeZone:   "local",
		TimeFormat: "2006-01-02T15:04:05-0700",
		Format:     logFormat(),
		Output:     os.Stdout,
	}))

	// wire up all route handlers
	handlers.Wire(r)

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

// logFormat returns the HTTP log format for the
// configured fiber logger middleware
func logFormat() string {
	if env.IsProd() {
		return logFormatProd
	}
	return logFormatDev
}

const logFormatProd = "${ip} ${header:x-forwarded-for} ${header:x-real-ip} " +
	"[${time}] ${pid} ${locals:requestid}${locals:lod-cache} \"${method} ${path} ${protocol}\" " +
	"${status} ${latency} ${bytesSent}b \"${referrer}\" \"${ua}\"\n"

const logFormatDev = "${ip} [${time}]${locals:lod-cache} \"${method} ${path} ${protocol}\" " +
	"${status} ${latency} ${bytesSent}b\n"
