package engine

import (
	"context"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

type ScreenshotResult struct {
	Buffer *[]byte
}

func performScreenshotTask(ctx context.Context, task *Task, logger *zerolog.Logger) (ScreenshotResult, error) {
	var buf []byte

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// The "full_page" task parameter captures the whole scrollable page
	// instead of just the viewport.
	capture := chromedp.CaptureScreenshot(&buf)
	if task.BoolParam("full_page", false) {
		capture = chromedp.FullScreenshot(&buf, 100)
	}

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(task.winWidth), int64(task.winHeight)),
		emulation.SetUserAgentOverride(task.userAgent),
		chromedp.Navigate(task.url),
		capture,
	)
	if err != nil {
		logger.Debug().Msgf("screenshot err: %v", err)

		return ScreenshotResult{}, err
	}

	return ScreenshotResult{&buf}, nil
}
