package engine

import (
	"context"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

type AnalyzeResult struct {
	Location          string            `json:"location"`
	RedirectLocations []Redirect        `json:"redirectLocations"`
	Body              string            `json:"body"`
	BodySize          int               `json:"bodySize"`
	InitialBody       string            `json:"initialBody"`
	InitialBodySize   int               `json:"initialBodySize"`
	AssetsCount       int               `json:"assetsCount"`
	Assets            map[string]*Asset `json:"assets"`
}

type Redirect struct {
	StatusCode int64  `json:"status_code"`
	Location   string `json:"location"`
}

type Asset struct {
	URL             string         `json:"url"`
	ResourceType    string         `json:"resourceType"`
	RequestHeaders  map[string]any `json:"requestHeaders"`
	ResponseHeaders map[string]any `json:"responseHeaders"`
	Body            string         `json:"body"`
}

func performAnalyzeTask(ctx context.Context, task *Task, logger *zerolog.Logger) (AnalyzeResult, error) {
	var mainReqID network.RequestID

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	result := AnalyzeResult{}
	result.Assets = make(map[string]*Asset)

	chromedp.ListenTarget(ctx, func(ev any) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			if ev.Initiator == nil {
				logger.Warn().Msg("analyze EventRequestWillBeSent with nil initiator")
				return
			}

			// "document" resource request types
			if ev.Type == network.ResourceTypeDocument && ev.Initiator.Type == "other" {
				if mainReqID == "" {
					mainReqID = ev.RequestID
				}

				// Capture redirects from navigation
				if ev.RedirectResponse != nil {
					if val, ok := ev.RedirectResponse.Headers["Location"]; ok {
						status := ev.RedirectResponse.Status

						switch location := val.(type) {
						case string:
							result.RedirectLocations = append(result.RedirectLocations, Redirect{status, location})
						}
					}
				}
			} else {
				// Track request as an asset
				result.Assets[string(ev.RequestID)] = &Asset{
					URL:            ev.Request.URL,
					ResourceType:   string(ev.Type),
					RequestHeaders: ev.Request.Headers,
				}
			}

		case *network.EventResponseReceived:
			if asset, ok := result.Assets[string(ev.RequestID)]; ok {
				asset.ResponseHeaders = ev.Response.Headers
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == mainReqID {
				go getResponseBody(ctx, ev.RequestID, func(body []byte, err error) {
					if err != nil {
						logger.Warn().Msgf("analyze getResponseBody main request erro: %s", err)
						return
					}

					result.InitialBody = string(body)
				})

				return
			} else if asset, ok := result.Assets[string(ev.RequestID)]; ok {
				go getResponseBody(ctx, ev.RequestID, func(body []byte, err error) {
					if err != nil {
						logger.Warn().Msgf("analyze getResponseBody main request erro: %s", err)
						return
					}

					asset.Body = string(body)
				})
			}
		}
	})

	initialSteps := []chromedp.Action{
		network.Enable(),
		chromedp.EmulateViewport(int64(task.winWidth), int64(task.winHeight)),
		emulation.SetUserAgentOverride(task.userAgent),
		chromedp.Navigate(task.url),
		chromedp.Location(&result.Location),
		chromedp.OuterHTML("html", &result.Body),
	}

	if err := chromedp.Run(ctx, initialSteps...); err != nil {
		return AnalyzeResult{}, err
	}

	// Set sizes and counts
	result.AssetsCount = len(result.Assets)
	result.InitialBodySize = len(result.InitialBody)
	result.BodySize = len(result.Body)

	return result, nil
}

func getResponseBody(ctx context.Context, reqID network.RequestID, callback func([]byte, error)) {
	var body []byte

	// ActionFunc to bind body and handle error
	fn := func(ctx context.Context) (err error) {
		body, err = network.GetResponseBody(reqID).Do(ctx)
		return err
	}

	err := chromedp.Run(ctx, chromedp.ActionFunc(fn))
	if err != nil {
		callback(body, err)
		return
	}

	callback(body, nil)
}
