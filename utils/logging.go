package utils

import (
	"strings"

	"github.com/rs/zerolog"
	zerologger "github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var Log zerolog.Logger

func InitLogging() {

	Log = zerologger.With().Logger().Level(logLevel(viper.GetString("log-level")))

}

func logLevel(input string) zerolog.Level {
	switch strings.ToLower(input) {
	case "none":
		return zerolog.Disabled
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "info", "information":
		return zerolog.InfoLevel
	case "err", "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerologger.Logger.GetLevel()
	}
}
