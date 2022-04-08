package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/dechristopher/lod/cache"
	"github.com/dechristopher/lod/config"
	"github.com/dechristopher/lod/env"
	"github.com/dechristopher/lod/str"
	"github.com/dechristopher/lod/util"
	"github.com/dechristopher/lod/www"
)

// main entry point to LOD
func main() {
	// set boot time immediately
	util.BootTime = time.Now()
	// print version info
	fmt.Printf(str.MInit, config.Version)

	// print message if in dev environment
	if !env.IsProd() {
		util.Info(str.CMain, str.MDevMode)
	}

	// parse and process command line flags
	parseFlags()

	// load .env if any
	_ = godotenv.Load()

	// read the config file
	if err := config.Load(); err != nil {
		if os.IsNotExist(err) {
			util.Error(str.CMain, str.EConfigNotFound, *config.File)
		} else {
			util.Error(str.CMain, str.EConfig, err.Error())
		}
		os.Exit(1)
	}

	// initialize cache instances
	if err := cache.Init(); err != nil {
		util.Error(str.CMain, str.EConfig, err.Error())
		os.Exit(1)
	}

	// serve LOD endpoints
	www.Serve()
}

// parseFlags parses and processes command line flags
// and shows the help message if requested
func parseFlags() {
	// parse command line flags
	config.File = flag.String(str.FConfigFile, "config.toml", str.FConfigFileUsage)
	env.IsDevFlag = flag.Bool(str.FDevMode, false, str.FDevModeUsage)
	util.DebugFlagPtr = flag.String(str.FDebugFlags, "", str.FDebugFlagsUsage)
	help := flag.Bool(str.FHelp, false, str.FHelpUsage)
	flag.Parse()

	// show help menu if "--help" is provided
	if *help {
		fmt.Printf(str.Help)
		os.Exit(0)
	}

	// parse out debug flags from command line options
	util.DebugFlags = strings.Split(*util.DebugFlagPtr, ",")
}
