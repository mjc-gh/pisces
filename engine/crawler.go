package engine

import (
	"context"
	"sync"
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

	// mu guards currentVisit, mainReqID, fetches and the contents of the
	// current visit. Network events are delivered on chromedp's event
	// goroutine while captureVisit runs on the task goroutine, and body
	// fetches complete on their own goroutines.
	mu           sync.Mutex
	currentVisit *Visit
	mainReqID    network.RequestID
	fetches      []chan struct{}
}

type Visit struct {
	RequestedURL      string           `json:"requested_url"`
	Location          string           `json:"location"`
	RedirectLocations []Redirect       `json:"redirect_locations"`
	CertificateInfo   *CertificateInfo `json:"certificate_info,omitempty"`
	Body              string           `json:"body"`
	InitialBody       string           `json:"initial_body,omitempty"`
	Assets            []*Asset         `json:"assets"`

	// assetsMap holds the assets of an in-progress visit, keyed by request ID.
	// assetOrder records the request IDs in the order they were first seen so
	// that Assets has a deterministic order.
	assetsMap  map[string]*Asset
	assetOrder []string
	// closed is set once the visit has been finalized; late network events and
	// body fetches no longer modify it.
	closed bool
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
		Visits:    make([]Visit, 0),
		userAgent: userAgent,
		winWidth:  winWidth,
		winHeight: winHeight,
	}
}

func (c *Crawler) SetupListeners(ctx context.Context, logger *zerolog.Logger) {
	chromedp.ListenTarget(ctx, func(ev any) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			c.onRequestWillBeSent(ev, logger)
		case *network.EventResponseReceived:
			c.onResponseReceived(ev)
		case *network.EventLoadingFinished:
			c.onLoadingFinished(ctx, ev, logger)
		}
	})
}

func (c *Crawler) onRequestWillBeSent(ev *network.EventRequestWillBeSent, logger *zerolog.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	visit := c.currentVisit
	if visit == nil {
		return
	}

	if ev.Initiator == nil {
		logger.Warn().Msg("crawler has nil initiator in EventRequestWillBeSent")

		return
	}

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

		return
	}

	id := string(ev.RequestID)
	if _, seen := visit.assetsMap[id]; !seen {
		visit.assetOrder = append(visit.assetOrder, id)
	}

	visit.assetsMap[id] = &Asset{
		URL:            ev.Request.URL,
		ResourceType:   string(ev.Type),
		RequestHeaders: ev.Request.Headers,
		InitiatorURL:   ev.Initiator.URL,
	}
}

func (c *Crawler) onResponseReceived(ev *network.EventResponseReceived) {
	c.mu.Lock()
	defer c.mu.Unlock()

	visit := c.currentVisit
	if visit == nil {
		return
	}

	secDetails := ev.Response.SecurityDetails

	if asset, ok := visit.assetsMap[string(ev.RequestID)]; ok {
		asset.CertificateInfo = getCertInfo(secDetails)
		asset.ResponseHeaders = ev.Response.Headers
		asset.ResponseStatus = ev.Response.Status
	} else if c.mainReqID == ev.RequestID {
		visit.CertificateInfo = getCertInfo(secDetails)
	}
}

func (c *Crawler) onLoadingFinished(ctx context.Context, ev *network.EventLoadingFinished, logger *zerolog.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	visit := c.currentVisit
	if visit == nil {
		return
	}

	if ev.RequestID == c.mainReqID {
		c.fetchBody(ctx, visit, ev.RequestID, func(body []byte, err error) {
			if err != nil {
				logger.Warn().Msgf("crawler getResponseBody main request error: %s", err)

				return
			}

			visit.InitialBody = string(body)
		})

		return
	}

	if asset, ok := visit.assetsMap[string(ev.RequestID)]; ok {
		c.fetchBody(ctx, visit, ev.RequestID, func(body []byte, err error) {
			if err != nil {
				logger.Warn().Msgf("crawler getResponseBody error: %s", err)

				return
			}

			asset.Body = string(body)
		})
	}
}

// fetchBody fetches a response body in the background and registers the fetch
// so that captureVisit can wait for it. The caller must hold c.mu. The apply
// callback is invoked with c.mu held, and only while the visit is still open.
func (c *Crawler) fetchBody(ctx context.Context, visit *Visit, reqID network.RequestID, apply func([]byte, error)) {
	done := make(chan struct{})
	c.fetches = append(c.fetches, done)

	go func() {
		defer close(done)

		getResponseBody(ctx, reqID, func(body []byte, err error) {
			c.mu.Lock()
			defer c.mu.Unlock()

			if visit.closed {
				return
			}

			apply(body, err)
		})
	}()
}

// waitForFetches blocks until every registered body fetch has finished, or ctx
// is done, and then closes the visit. The visit is closed in the same critical
// section that observes no pending fetches, so a fetch can not be registered
// and missed. Fetches still running when ctx is done are discarded.
func (c *Crawler) waitForFetches(ctx context.Context, visit *Visit) {
	for {
		c.mu.Lock()

		pending := c.fetches
		c.fetches = nil

		if len(pending) == 0 || ctx.Err() != nil {
			visit.closed = true
			c.mu.Unlock()

			return
		}

		c.mu.Unlock()

		for _, done := range pending {
			select {
			case <-done:
			case <-ctx.Done():
			}
		}
	}
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

// captureVisit runs callback while recording the network activity of the page
// into a new Visit. Once the callback returns it waits for every in-flight body
// fetch to finish (or ctx to be done) before assembling the Visit, so the
// result is complete and no longer touched by background goroutines.
func (c *Crawler) captureVisit(ctx context.Context, callback visitCallback) error {
	visit := Visit{}
	visit.RedirectLocations = []Redirect{}
	visit.assetsMap = make(map[string]*Asset)

	c.mu.Lock()
	c.currentVisit = &visit
	c.mu.Unlock()

	if err := callback(ctx); err != nil {
		c.abortVisit(&visit)

		return err
	}

	c.waitForFetches(ctx, &visit)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Flatten the assets map to a slice of Assets, in the order first seen.
	for _, id := range visit.assetOrder {
		visit.Assets = append(visit.Assets, visit.assetsMap[id])
	}

	// Clear the working state (no longer needed)
	visit.assetsMap = nil
	visit.assetOrder = nil

	// Update crawler state
	c.Visits = append(c.Visits, visit)
	c.currentVisit = nil
	c.mainReqID = ""
	c.fetches = nil

	return nil
}

// abortVisit discards the visit in progress after a failure.
func (c *Crawler) abortVisit(visit *Visit) {
	c.mu.Lock()
	defer c.mu.Unlock()

	visit.closed = true
	c.currentVisit = nil
	c.mainReqID = ""
	c.fetches = nil
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
