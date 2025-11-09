package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

const (
	CHROME_DP_HEADLESS_SHELL = "http://127.0.0.1:9222/json/version"
)

func StartLocal(ctx context.Context) (context.Context, context.CancelFunc) {
	return chromedp.NewExecAllocator(ctx, chromedp.Headless)
}

func StartRemote(ctx context.Context, url string) (context.Context, context.CancelFunc) {
	return chromedp.NewRemoteAllocator(ctx, url)
}
