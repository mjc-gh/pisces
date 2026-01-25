package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/engine"
	"github.com/mjc-gh/pisces/internal/rest"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

var (
	version string
)

func main() {
	ver := version
	if ver == "" {
		ver = "0.0.0"
	}

	cmd := &cli.Command{
		Name:    "pisces-web",
		Version: ver,
		Usage:   "A web server API for analyzing phishing sites.",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "debug", Aliases: []string{"d"}, Usage: "enable debug logging"},
			&cli.StringFlag{Name: "host", Value: "127.0.0.1", Usage: "web server bind host"},
			&cli.IntFlag{Name: "port", Value: 8888, Usage: "web server bind port"},
			&cli.IntFlag{Name: "remote-port", Usage: "remote DevTools port"},
			&cli.StringFlag{Name: "remote-host", Usage: "remote DevTools host"},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			ctx, cancel := context.WithCancel(c)

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGTERM)
			// signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

			remoteHost := cmd.String("remote-host")
			remotePort := cmd.Int("remote-port")

			host := cmd.String("host")
			port := cmd.Int("port")
			if port == 0 {
				port = 8888
			}

			logger := pisces.SetupLogger(cmd.Bool("debug")).
				With().
				Str("service", "pisces-web").
				Str("source", "go").
				Logger()

			go func(done context.CancelFunc, l *zerolog.Logger) {
				// TODO: stop receiving new task requests

				sig := <-signalChan

				l.Info().Msgf("signal %v received; shutting down", sig)

				// TODO: give the engine more time to finish work before cancelling?
				done()
			}(cancel, &logger)

			opts := []engine.Option{engine.WithLogger(&logger)}

			if remoteHost != "" && remotePort != 0 {
				opts = append(opts, engine.WithRemoteAllocator(remoteHost, remotePort))
			} else if cmd.Bool("headfull") {
				opts = append(opts, engine.WithHeadfullLocalAllocator())
			}

			e := engine.New(cmd.Int("concurrency"), opts...)
			e.Start(ctx)

			addr := fmt.Sprintf("%s:%d", host, port)
			if err := rest.StartServer(ver, addr, e, &logger); err != nil {
				logger.Warn().Msgf("failed to start REST web server: %v", err)

				os.Exit(1)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
