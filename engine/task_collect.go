package engine

import (
	"context"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

type CollectResult struct {
	BodyLength        int `json:"body_length"`
	InitialBodyLength int `json:"initial_body_length"`
	TotalAssets       int `json:"total_assets"`
	*Visit
}

func performCollectTask(ctx context.Context, task *Task, logger *zerolog.Logger) (CollectResult, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	crawler := NewCrawler(task.userAgent, int64(task.winWidth), int64(task.winHeight))
	err := crawler.Visit(ctx, task.url, logger)
	if err != nil {
		return CollectResult{}, err
	}

	visit := crawler.LastVisit()
	if visit == nil {
		return CollectResult{}, ErrNoCrawlerVisit
	}

	result := CollectResult{Visit: visit}

	// Set result metadata
	result.InitialBodyLength = len(result.InitialBody)
	result.BodyLength = len(result.Body)
	result.TotalAssets = len(result.Assets)

	return result, nil
}
