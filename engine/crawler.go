package engine

import (
	"context"
	"time"

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

	currentVisit *Visit
	mainReqID    network.RequestID
}

type Visit struct {
	RequestedURL      string           `json:"requested_url"`
	Location          string           `json:"location"`
	RedirectLocations []Redirect       `json:"redirect_locations"`
	CertificateInfo   *CertificateInfo `json:"certificate_info,omitempty"`
	Body              string           `json:"body"`
	InitialBody       string           `json:"initial_body,omitempty"`
	Assets            []*Asset         `json:"assets"`

	assetsMap map[string]*Asset
}

type Redirect struct {
	StatusCode int64  `json:"status_code"`
	Location   string `json:"location"`
}

type Asset struct {
	URL             string           `json:"url"`
	CertificateInfo *CertificateInfo `json:"certificate_info,omitempty"`
	ResourceType    string           `json:"resource_type"`
	RequestHeaders  map[string]any   `json:"request_headers"`
	ResponseHeaders map[string]any   `json:"response_headers"`
	ResponseStatus  int64            `json:"response_status,omitempty"`
	InitiatorURL    string           `json:"initiator_url,omitempty"`
	Body            string           `json:"body,omitempty"`
}

type CertificateInfo struct {
	Protocol    string    `json:"protocol"`
	Issuer      string    `json:"issuers"`
	SubjectName string    `json:"subject_name"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to"`
	SANs        []string  `json:"sans,omitempty"`
}

func NewCrawler(userAgent string, winWidth, winHeight int64) Crawler {
	return Crawler{
		make([]Visit, 0), userAgent, winWidth, winHeight, nil, "",
	}
}

func (c *Crawler) SetupListeners(ctx context.Context, logger *zerolog.Logger) {
	chromedp.ListenTarget(ctx, func(ev any) {
		visit := c.currentVisit
		if visit == nil {
			return
		}

		// case *page.EventLifecycleEvent:
		//   Name == "networkIdle"

		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			// log.Println("EventResponseReceived")

			if ev.Initiator == nil {
				logger.Warn().Msg("crawler has nil initiator in EventRequestWillBeSent")

				return
			}

			// "document" resource request types
			// log.Printf("%s %s", ev.Type, ev.Initiator.Type)
			if ev.Type == network.ResourceTypeDocument && ev.Initiator.Type == "other" {
				if c.mainReqID == "" {
					c.mainReqID = ev.RequestID
				}

				// Capture redirects from navigation
				if ev.RedirectResponse != nil {
					if val, ok := ev.RedirectResponse.Headers["Location"]; ok {
						status := ev.RedirectResponse.Status

						switch location := val.(type) {
						case string:
							visit.RedirectLocations = append(visit.RedirectLocations, Redirect{status, location})
						default:
							logger.Debug().Msgf("redirect location capture unexpected type: %v", val)
						}
					}
				}
			} else {
				// Track request as an asset
				// log.Printf("assetsMap Request URL %v\n", ev.Request.URL)
				visit.assetsMap[string(ev.RequestID)] = &Asset{
					URL:            ev.Request.URL,
					ResourceType:   string(ev.Type),
					RequestHeaders: ev.Request.Headers,
					InitiatorURL:   ev.Initiator.URL,
				}
			}

		case *network.EventResponseReceived:
			// log.Println("EventResponseReceived")

			secDetails := ev.Response.SecurityDetails

			if asset, ok := visit.assetsMap[string(ev.RequestID)]; ok {
				asset.CertificateInfo = getCertInfo(secDetails)
				asset.ResponseHeaders = ev.Response.Headers
				asset.ResponseStatus = ev.Response.Status
			} else if c.mainReqID == ev.RequestID {
				visit.CertificateInfo = getCertInfo(secDetails)
			}
		case *network.EventLoadingFinished:
			// log.Println("EventLoadingFinished")

			if ev.RequestID == c.mainReqID {
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
			// default:
			// log.Printf("%T\n", ev)
		}
	})
}

// Visit a given URL and capture all the assets the document includes.
func (c *Crawler) Visit(ctx context.Context, url string, logger *zerolog.Logger) error {
	return c.captureVisit(ctx, func(ctx context.Context) error {
		visit := c.currentVisit
		if visit == nil {
			return ErrInvalidCrawlerState
		}

		visit.RequestedURL = url

		visitSteps := chromedp.Tasks{
			network.Enable(),
			chromedp.EmulateViewport(c.winWidth, c.winHeight),
			emulation.SetUserAgentOverride(c.userAgent),
			chromedp.Navigate(url),
			chromedp.Location(&visit.Location),
			chromedp.OuterHTML("html", &visit.Body),
		}

		return chromedp.Run(ctx, visitSteps...)
	})
}

func (c *Crawler) LastVisit() *Visit {
	l := len(c.Visits)
	if l < 1 {
		return nil
	}

	return &c.Visits[l-1]
}

type visitCallback func(context.Context) error

// Visit a given URL and capture all the assets the document includes.
func (c *Crawler) captureVisit(ctx context.Context, callback visitCallback) error {
	visit := Visit{}
	visit.RedirectLocations = []Redirect{}
	visit.assetsMap = make(map[string]*Asset)

	c.currentVisit = &visit

	if err := callback(ctx); err != nil {
		return err
	}

	// Flatten the assets map to slice of Assets
	for _, asset := range visit.assetsMap {
		visit.Assets = append(visit.Assets, asset)
	}

	// Clear assetMap (no longer needed)
	visit.assetsMap = nil

	// Update crawler state
	c.Visits = append(c.Visits, visit)
	c.currentVisit = nil
	c.mainReqID = ""

	return nil
}

func getResponseBody(ctx context.Context, reqID network.RequestID, callback func([]byte, error)) {
	var body []byte

	// ActionFunc to bind body and handle error
	fn := func(ctx context.Context) error {
		var err error

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

func getCertInfo(secDetails *network.SecurityDetails) *CertificateInfo {
	if secDetails != nil && secDetails.Protocol != "" && secDetails.Issuer != "" {
		return &CertificateInfo{
			secDetails.Protocol,
			secDetails.Issuer,
			secDetails.SubjectName,
			secDetails.ValidFrom.Time(),
			secDetails.ValidTo.Time(),
			secDetails.SanList,
		}
	}

	return nil
}
