package engine

import (
	"context"
	"errors"

	"github.com/chromedp/chromedp"
	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/rs/zerolog"
)

type CollectResult struct {
	*browser.Visit
	AssetsCount     int
	BodySize        int
	InitialBodySize int
}

func performCollectTask(ctx context.Context, task *Task, logger *zerolog.Logger) (CollectResult, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	crawler := browser.NewCrawler(task.userAgent, int64(task.winWidth), int64(task.winHeight))
	err := crawler.Visit(ctx, task.url, logger)
	if err != nil {
		return CollectResult{}, err
	}

	visit := crawler.LastVisit()
	if visit == nil {
		return CollectResult{}, errors.New("no visit from crawler")
	}

	result := CollectResult{Visit: visit}

	// Set sizes and counts
	result.AssetsCount = len(result.Assets)
	result.InitialBodySize = len(result.InitialBody)
	result.BodySize = len(result.Body)

	return result, nil
}
