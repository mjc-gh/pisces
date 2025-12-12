package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/engine"
	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

var ErrInvalidDeviceProperties = errors.New("invalid device properties")
var ErrScreenShotFailed = errors.New("screenshot result error")

var (
	logger  *zerolog.Logger
	version string
)

type taskCallbackFn = func(*cli.Command, *engine.Engine) error

func main() {
	baseArgs := []cli.Argument{
		&cli.StringArgs{Name: "url", Min: 1, Max: -1},
	}

	baseFlags := []cli.Flag{
		&cli.BoolFlag{Name: "debug", Aliases: []string{"d"}, Usage: "enable debug logging"},
		&cli.BoolFlag{Name: "remote", Aliases: []string{"r"}, Usage: "use a remote Chrome DevTools instance"},
		&cli.BoolFlag{Name: "headfull", Aliases: []string{"H"}, Usage: "run browser in headfull mode"},
		&cli.IntFlag{Name: "concurrency", Aliases: []string{"c"}, Usage: "number of concurrent workers"},
		&cli.IntFlag{Name: "port", Value: 9222, Usage: "remote DevTools port"},
		&cli.StringFlag{Name: "device-type", Value: "desktop", Usage: "device type (desktop/mobile/tablet)", Action: validDeviceType},
		&cli.StringFlag{Name: "device-size", Value: "large", Usage: "device size preset", Action: validDeviceSize},
		&cli.StringFlag{Name: "host", Value: "127.0.0.1", Usage: "remote DevTools host"},
		&cli.StringFlag{Name: "user-agent", Value: "chrome", Usage: "browser user-agent preset"},
		&cli.StringFlag{
			Name:  "rules-dir",
			Value: "rules",
			Usage: "directory containing Sigma rules (default: rules/)",
		},
	}

	withOutputFlags := append([]cli.Flag{
		&cli.StringFlag{Name: "output", Value: "pisces.ndjson", Aliases: []string{"o"}, Usage: "output NDJSON file"},
	}, baseFlags...)

	withWaitFlags := append([]cli.Flag{
		&cli.IntFlag{Name: "wait", Value: 300, Aliases: []string{"w"}, Usage: "wait time (seconds) after load before analysis"},
	}, withOutputFlags...)

	ver := version
	if ver == "" {
		ver = "0.0.0"
	}

	cmd := &cli.Command{
		Name:    "pisces",
		Version: ver,
		Usage:   "A tool for analyzing phishing sites",
		Commands: []*cli.Command{
			{
				Name:      "analyze",
				Usage:     "Analyze and interact one or more URLs for phishing",
				Arguments: baseArgs,
				Flags:     withWaitFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					params := map[string]any{
						"wait": cmd.Int("wait"),
					}

					return runTask(ctx, cmd, "analyze", params, outputResultJson)
				},
			},
			{
				Name:      "collect",
				Usage:     "Collect HTML and assets for one or more URLs",
				Arguments: baseArgs,
				Flags:     withOutputFlags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runTask(ctx, cmd, "collect", map[string]any{}, outputResultJson)
				},
			},
			{
				Name:      "screenshot",
				Usage:     "Screenshot one or more URLs",
				Arguments: baseArgs,
				Flags: append([]cli.Flag{
					&cli.StringFlag{Name: "output-dir", Value: "tmp/", Aliases: []string{"o"}, Usage: "directory for screenshots"},
				}, baseFlags...),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runTask(ctx, cmd, "screenshot", map[string]any{}, screenshotCallback)
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runTask(ctx context.Context, cmd *cli.Command, name string, params map[string]any, callback taskCallbackFn) error {
	logger = pisces.SetupLogger(cmd.Bool("debug"))

	deviceSize := cmd.StringArg("device-size")
	deviceType := cmd.StringArg("device-type")
	host := cmd.String("host")
	port := cmd.Int("port")
	urls := cmd.StringArgs("url")

	opts := []engine.Option{engine.WithLogger(pisces.Logger())}

	if cmd.Bool("remote") && host != "" && port != 0 {
		opts = append(opts, engine.WithRemoteAllocator(host, port))
	} else if cmd.Bool("headfull") {
		opts = append(opts, engine.WithHeadfullLocalAllocator())
	}

	e := engine.New(cmd.Int("concurrency"), opts...)
	e.Start(ctx)

	for _, url := range urls {
		t := engine.NewTask(
			name, url,
			engine.WithParams(params),
			engine.WithDeviceProperties(deviceType, deviceSize),
			engine.WithUserAgent(deviceType, cmd.StringArg("user-agent")),
		)

		e.Add(t)
	}

	go e.Shutdown()

	// TODO handle interrupt signal and wait for shutdown

	return callback(cmd, e)
}

func outputResultJson(cmd *cli.Command, e *engine.Engine) error {
	output := cmd.String("output")
	out, err := os.Create(filepath.Clean(output))
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		cerr := out.Close()
		if cerr != nil {
			logger.Warn().Err(cerr).Msg("file close error")
		}
	}()

	rulesDir := cmd.String("rules-dir")
	if rulesDir == "" {
		rulesDir = "rules"
	}

	if rulesDir != "rules" {
		logger.Info().Str("rules_dir", rulesDir).Msg("using custom Sigma rules directory")
	}

	if err := engine.InitSigmaEngine(rulesDir, logger); err != nil {
		logger.Warn().Err(err).Msg("failed to initialize Sigma engine")
	}

	ctx := context.Background()

	for r := range e.Results() {
		if r.Error != nil {
			logger.Warn().Msgf("result error: %v", r.Error)

			continue
		}

		matches, err := engine.EvaluateSigmaResult(ctx, r, logger)
		if err != nil {
			logger.Warn().Err(err).Msg("sigma evaluation error")
		} else {
			for _, m := range matches {
				if !m.Match {
					continue
				}
				logger.Info().
					Str("rule_id", m.Rule.ID).
					Str("title", m.Rule.Title).
					Str("url", r.URL).
					Msg("[+] SIGMA ALERT: rule matched for URL")
			}
		}

		logger.Info().
			Str("url", r.URL).
			Str("duration", r.Elapsed.String()).
			Msg("result")

		line, err := json.Marshal(r.Result)
		if err != nil {
			logger.Warn().Err(err).Msg("result JSON marshal error")

			continue
		}
		if _, err = out.Write(line); err != nil {
			logger.Warn().Err(err).Msg("result JSON write error")

			continue
		}
		if _, err = out.WriteString("\n"); err != nil {
			logger.Warn().Err(err).Msg("result JSON newline write error")

			continue
		}
		logger.Debug().Msgf("wrote to file %s", output)
	}

	logger.Info().Msg("done")

	return nil
}

func screenshotCallback(cmd *cli.Command, e *engine.Engine) error {
	outputDir := cmd.String("output-dir")
	err := os.MkdirAll(outputDir, 0o750)
	if err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	for r := range e.Results() {
		if r.Error != nil {
			logger.Warn().Msgf("result error: %v", r.Error)

			continue
		}

		logger.Info().
			Str("url", r.URL).
			Str("duration", r.Elapsed.String()).
			Msg("screenshot result")

		sr, ok := r.Result.(*engine.ScreenshotResult)
		if !ok {
			return ErrScreenShotFailed
		}

		fileName, err := urlToFilename(r.URL)
		if err != nil {
			return fmt.Errorf("build screenshot filename: %w", err)
		}

		outPath := filepath.Join(outputDir, fileName+".png")
		out, err := os.Create(filepath.Clean(outPath))
		if err != nil {
			return fmt.Errorf("create screenshot file: %w", err)
		}

		if _, err = out.Write(*sr.Buffer); err != nil {
			// Try to close file, but ignore any errors
			_ = out.Close()

			return fmt.Errorf("write screenshot file: %w", err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("close screenshot file: %w", err)
		}
	}

	return nil
}

func validDeviceType(ctx context.Context, cmd *cli.Command, v string) error {
	if !browser.IsValidDeviceType(v) {
		return fmt.Errorf("%w: %v", ErrInvalidDeviceProperties, v)
	}

	return nil
}

func validDeviceSize(ctx context.Context, cmd *cli.Command, v string) error {
	if !browser.IsValidDeviceSize(v) {
		return fmt.Errorf("%w: %v", ErrInvalidDeviceProperties, v)
	}

	return nil
}

func urlToFilename(taskURL string) (string, error) {
	u, err := url.Parse(taskURL)
	if err != nil {
		return "", err
	}

	domain := u.Host
	path := u.Path

	combined := strings.Trim(domain+path, "/")

	safe := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '.':
			return '_'
		default:
			return r
		}
	}, combined)

	if safe == "" {
		safe = "index"
	}

	return safe, nil
}
