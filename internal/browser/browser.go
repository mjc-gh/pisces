package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

// StartLocal creates a local exec allocator. Any extra options are appended
// after the defaults so that they take precedence.
func StartLocal(ctx context.Context, headfull bool, extra ...chromedp.ExecAllocatorOption) (context.Context, context.CancelFunc) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("block-new-web-contents", true),
	}

	if !headfull {
		opts = append(opts, chromedp.Headless)
	}

	opts = append(opts, extra...)

	return chromedp.NewExecAllocator(ctx, opts...)
}

func StartRemote(ctx context.Context, url string) (context.Context, context.CancelFunc) {
	return chromedp.NewRemoteAllocator(ctx, url)
}
