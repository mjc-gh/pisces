package browser

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
	Location          string            `json:"location"`
	RedirectLocations []Redirect        `json:"redirectLocations"`
	Body              string            `json:"body"`
	InitialBody       string            `json:"initialBody"`
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

func NewCrawler(userAgent string, winWidth, winHeight int64) Crawler {
	return Crawler{
		make([]Visit, 1), userAgent, winWidth, winHeight,
	}
}

func (c *Crawler) Visit(ctx context.Context, url string, logger *zerolog.Logger) error {
	var mainReqID network.RequestID

	visit := Visit{}
	visit.Assets = make(map[string]*Asset)

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
				visit.Assets[string(ev.RequestID)] = &Asset{
					URL:            ev.Request.URL,
					ResourceType:   string(ev.Type),
					RequestHeaders: ev.Request.Headers,
				}
			}

		case *network.EventResponseReceived:
			if asset, ok := visit.Assets[string(ev.RequestID)]; ok {
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
			} else if asset, ok := visit.Assets[string(ev.RequestID)]; ok {
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

	initialSteps := []chromedp.Action{
		network.Enable(),
		chromedp.EmulateViewport(int64(c.winWidth), int64(c.winHeight)),
		emulation.SetUserAgentOverride(c.userAgent),
		chromedp.Navigate(url),
		chromedp.Location(&visit.Location),
		chromedp.OuterHTML("html", &visit.Body),
	}

	if err := chromedp.Run(ctx, initialSteps...); err != nil {
		return err
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
