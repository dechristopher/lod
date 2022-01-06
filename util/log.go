package util

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tile-fund/lod/env"
	"github.com/tile-fund/lod/str"
)

// LogMessage is a json struct used for non-stdout log messages
type LogMessage struct {
	Time     int64  `json:"time"`
	Severity string `json:"severity"`
	Caller   string `json:"caller"`
	Message  string `json:"message"`
}

// Info prints an info message to the standard logger
func Info(caller, message string, args ...interface{}) {
	printLog("info", str.InfoFormat, caller, message, args...)
}

// Debug prints a debug message to the standard logger
func Debug(caller, message string, args ...interface{}) {
	if env.IsDev() {
		printLog("debug", str.DebugFormat, caller, message, args...)
	}
}

// DebugFlag prints a debug message to the standard logger if flag is enabled
func DebugFlag(flag, caller, message string, args ...interface{}) {
	if IsDebugFlag(flag) {
		Debug(caller, message, args...)
	}
}

// Error prints an error message to the standard logger
func Error(caller, message string, args ...interface{}) {
	printLog("warn", str.ErrorFormat, caller, message, args...)
}

// printLog prints logs to stdout in the proper format
// Standard in developer mode and JSON in deploy mode
func printLog(severity, format, caller, message string, args ...interface{}) {
	if env.GetEnv() == env.Dev {
		log.Printf(format, caller, fmt.Sprintf(message, args...))
	} else {
		logMessage := LogMessage{
			Time:     MilliTime(),
			Severity: severity,
			Caller:   strings.TrimSpace(caller),
			Message:  fmt.Sprintf(message, args...),
		}
		if out, err := json.Marshal(logMessage); err != nil {
			Error(str.CLog, str.ELogFail, err.Error(), logMessage)
		} else {
			fmt.Printf("%s\n", out)
		}
	}
}
