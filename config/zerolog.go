package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoadLog - setting format, level and std of zerolog
func LoadLog() {

	switch os.Getenv("ZERO_LOG_LEVEL") {
	case "PANIC":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "NOLEVEL":
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	case "DISABLED":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	ztf, ok := os.LookupEnv("ZERO_LOG_TIME_FORMAT")
	if !ok {
		log.Print("No log time format in Env default \"2006-01-02 15:04:05.000\" taken")
		ztf = "2006-01-02 15:04:05.000"
	}
	zerolog.TimeFieldFormat = ztf

}
