package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	Debug     *log.Logger
	Info      *log.Logger
	Warning   *log.Logger
	Error     *log.Logger
	debugMode bool
)

// Init initializes loggers.
// Debug logs are controlled by ADMINA_DEBUG environment variable.
func Init() {
	debugMode = os.Getenv("ADMINA_DEBUG") == "true"
	debugHandle := io.Discard
	if debugMode {
		debugHandle = os.Stderr
	}

	Debug = log.New(debugHandle,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(os.Stderr,
		"INFO: ",
		log.Ldate|log.Ltime)

	Warning = log.New(os.Stderr,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// LogDebug outputs debug log if debug mode is enabled
func LogDebug(format string, args ...interface{}) {
	if debugMode {
		Debug.Printf(format, args...)
	}
}

// LogInfo outputs info log
func LogInfo(format string, args ...interface{}) {
	Info.Printf(format, args...)
}

// LogWarning outputs warning log
func LogWarning(format string, args ...interface{}) {
	Warning.Printf(format, args...)
}

// LogError outputs error log
func LogError(format string, args ...interface{}) {
	Error.Printf(format, args...)
}

// Print outputs to stdout without log formatting
func Print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// PrintErr outputs to stderr without log formatting
func PrintErr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}
