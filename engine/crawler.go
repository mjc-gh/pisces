package engine

import (
	"context"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
)

type Crawler struct {
	Visits    []Visit
	userAgent string
	winWidth  int64
	winHeight int64
}

type Visit struct {
	RequestedUrl      string     `json:"requested_url"`
	Location          string     `json:"location"`
	RedirectLocations []Redirect `json:"redirect_locations"`
	Body              string     `json:"body"`
	InitialBody       string     `json:"initial_body"`
	Assets            []*Asset   `json:"assets"`

	assetsMap map[string]*Asset
}

type Redirect struct {
	StatusCode int64  `json:"status_code"`
	Location   string `json:"location"`
}

type Asset struct {
	URL             string         `json:"url"`
	ResourceType    string         `json:"resource_type"`
	RequestHeaders  map[string]any `json:"request_headers"`
	ResponseHeaders map[string]any `json:"response_headers"`
	InitiatorURL    string         `json:"initiator_url,omitempty"`
	Body            string         `json:"body,omitempty"`
}

func NewCrawler(userAgent string, winWidth, winHeight int64) Crawler {
	return Crawler{
		make([]Visit, 0), userAgent, winWidth, winHeight,
	}
}

func (c *Crawler) Visit(ctx context.Context, url string, logger *zerolog.Logger) error {
	var mainReqID network.RequestID

	visit := Visit{RequestedUrl: url}
	visit.assetsMap = make(map[string]*Asset)

	chromedp.ListenTarget(ctx, func(ev any) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			if ev.Initiator == nil {
				logger.Warn().Msg("crawler has nil initiator in EventRequestWillBeSent")
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
							visit.RedirectLocations = append(visit.RedirectLocations, Redirect{status, location})
						}
					}
				}
			} else {
				// Track request as an asset
				visit.assetsMap[string(ev.RequestID)] = &Asset{
					URL:            ev.Request.URL,
					ResourceType:   string(ev.Type),
					RequestHeaders: ev.Request.Headers,
					InitiatorURL:   ev.Initiator.URL,
				}
			}

		case *network.EventResponseReceived:
			if asset, ok := visit.assetsMap[string(ev.RequestID)]; ok {
				asset.ResponseHeaders = ev.Response.Headers
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == mainReqID {
				go getResponseBody(ctx, ev.RequestID, func(body []byte, err error) {
					if err != nil {
						logger.Warn().Msgf("crawler getResponseBody main request error: %s", err)
						return
					}

					visit.InitialBody = string(body)
				})

				return
			} else if asset, ok := visit.assetsMap[string(ev.RequestID)]; ok {
				go getResponseBody(ctx, ev.RequestID, func(body []byte, err error) {
					if err != nil {
						logger.Warn().Msgf("crawler getResponseBody error: %s", err)
						return
					}

					asset.Body = string(body)
				})
			}
		}
	})

	visitSteps := []chromedp.Action{
		network.Enable(),
		chromedp.EmulateViewport(int64(c.winWidth), int64(c.winHeight)),
		emulation.SetUserAgentOverride(c.userAgent),
		chromedp.Navigate(url),
		chromedp.Location(&visit.Location),
		chromedp.OuterHTML("html", &visit.Body),
	}

	if err := chromedp.Run(ctx, visitSteps...); err != nil {
		return err
	}

	// Flatten the assets map to slice of Assets
	for _, asset := range visit.assetsMap {
		visit.Assets = append(visit.Assets, asset)
	}

	// Add visit to slice
	c.Visits = append(c.Visits, visit)

	return nil
}

func (c *Crawler) LastVisit() *Visit {
	l := len(c.Visits)
	if l < 1 {
		return nil
	}

	return &c.Visits[l-1]
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
