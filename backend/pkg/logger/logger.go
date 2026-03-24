package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the global structured logger instance.
var Log zerolog.Logger

func init() {
	if os.Getenv("GIN_MODE") == "release" {
		Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}).
			With().Timestamp().Logger()
	}
}
