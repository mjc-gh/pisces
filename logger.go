package pisces

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger *zerolog.Logger

func SetupLogger(debug bool) *zerolog.Logger {

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	l := zerolog.New(os.Stdout).With().Timestamp().
		Str("source", "go").
		Str("service", "pisces").
		Logger().
		Sample(
			zerolog.LevelSampler{
				TraceSampler: &zerolog.BurstSampler{
					Burst:       1,
					Period:      2 * time.Second,
					NextSampler: &zerolog.BasicSampler{N: 100},
				},
				WarnSampler: &zerolog.BurstSampler{
					Burst:       4,
					Period:      1 * time.Second,
					NextSampler: &zerolog.BasicSampler{N: 100},
				},
			},
		)

	logger = &l

	return logger
}

func Logger() *zerolog.Logger {
	if logger == nil {
		nop := zerolog.Nop()

		return &nop
	}

	return logger
}
