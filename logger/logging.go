// logger should immediately be replaced by https://pkg.go.dev/log/slog@master
// when the project upgrades to Go 1.21
package logger

import (
	"log"
	"os"
)

var debug = false

func init() {
	if os.Getenv("FAAS_DEBUG") == "1" {
		debug = true
	}
}

func Debug(message string) {
	if debug {
		log.Println(message)
	}
}

func Debugf(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}

func Print(message string) {
	log.Println(message)
}

func Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
