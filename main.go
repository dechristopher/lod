package main

import (
	"flag"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
	"github.com/tile-fund/lod/www"
)

// init parses flags, sets constants, and prepares us for battle
func init() {
	// set boot time immediately
	util.BootTime = time.Now()

	// parse command line flags
	util.DebugFlagPtr = flag.String(str.FDebugFlags, "", str.FDebugFlagsUsage)
	flag.Parse()

	// parse out debug flags from command line options
	util.DebugFlags = strings.Split(*util.DebugFlagPtr, ",")

	if env.IsDev() {
		util.Info(str.CMain, str.MDevMode)
	}
}

// main does the things
func main() {
	// load .env if any
	_ = godotenv.Load()

	// serve LOD endpoints
	www.Serve()
}
