package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mjc-gh/pisces/engine"
	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:  "analyze",
				Usage: "Analyze one or more URLs",
				Arguments: []cli.Argument{
					&cli.StringArgs{Name: "url", Min: 0, Max: -1},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "debug", Aliases: []string{"d"}},
					&cli.BoolFlag{Name: "remote", Aliases: []string{"r"}},
					&cli.IntFlag{Name: "concurrency", Aliases: []string{"c"}},
					&cli.IntFlag{Name: "port", Value: 9222},
					&cli.StringFlag{Name: "device-type", Value: "desktop", Action: validDeviceType},
					&cli.StringFlag{Name: "device-size", Value: "large", Action: validDeviceSize},
					&cli.StringFlag{Name: "host", Value: "127.0.0.1"},
					&cli.StringFlag{Name: "output", Value: "pisces.json", Aliases: []string{"o"}},
					&cli.StringFlag{Name: "user-agent", Value: "chrome"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					logger := engine.NewLogger(cmd.Bool("debug"))

					deviceSize := cmd.StringArg("device-size")
					deviceType := cmd.StringArg("device-type")
					host := cmd.String("host")
					output := cmd.String("output")
					port := cmd.Int("port")
					urls := cmd.StringArgs("url")

					opts := []engine.Option{engine.WithLogger(logger)}

					if cmd.Bool("remote") && host != "" && port != 0 {
						opts = append(opts, engine.WithRemoteAllocator(host, port))
					}

					e := engine.New(cmd.Int("concurrency"), opts...)
					e.Start(ctx)

					for _, url := range urls {
						t := engine.NewTask("analyze", url)
						t.SetDevice(deviceType, deviceSize)
						t.SetUserAgent(deviceType, cmd.StringArg("user-agent"))

						e.Add(t)
					}

					go e.Shutdown()

					// TODO handle interrupt signal and wait for shutdown

					out, err := os.Create(output)
					if err != nil {
						panic(err)
					}

					defer out.Close()

					// TODO write an arary for results as JSON
					for r := range e.Results() {
						if r.Error != nil {
							logger.Warn().Msgf("result error: %v", r.Error)
							continue
						}

						logger.Info().Msgf("result for %s (duration %s)", r.URL, r.Elapsed.String())

						json, err := r.JSON()
						if err != nil {
							logger.Warn().Msgf("result json error: %v", err)
							continue
						}

						_, err = out.Write(json)
						if err != nil {
							logger.Warn().Msgf("result json write error: %v", err)
						}

						logger.Info().Msgf("wrote to file %s", output)
					}

					logger.Info().Msg("done")

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func validDeviceType(ctx context.Context, cmd *cli.Command, v string) error {
	if !browser.IsValidDeviceType(v) {
		return fmt.Errorf("invalid device type: %v", v)
	}

	return nil
}

func validDeviceSize(ctx context.Context, cmd *cli.Command, v string) error {
	if !browser.IsValidDeviceSize(v) {
		return fmt.Errorf("invalid device size: %v", v)
	}

	return nil
}
