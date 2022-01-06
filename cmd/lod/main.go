package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tile-fund/lod/config"

	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
	"github.com/tile-fund/lod/util"
	"github.com/tile-fund/lod/www"
)

// init parses flags, sets constants, and prepares us for battle
func init() {
	// set boot time immediately
	util.BootTime = time.Now()

	// print version info
	fmt.Printf(str.MInit, config.Version)

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

	// read the config file
	if err := config.Read(); err != nil {
		if os.IsNotExist(err) {
			util.Error(str.CMain, str.EConfigNotFound, *config.File)
		} else {
			util.Error(str.CMain, str.EConfig, err.Error())
		}
		os.Exit(1)
	}

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
