package engine

import (
	"slices"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func matchAsset(fileName string) func(*Asset) bool {
	return func(a *Asset) bool {
		return strings.Contains(a.URL, fileName)
	}
}

func TestCrawlerVisit(t *testing.T) {
	pctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ctx, _ := chromedp.NewContext(pctx)
	server := piscestest.NewTestWebServer("simple")
	defer server.Close()
	crawler := NewCrawler("pisces", 1920, 1080)
	crawler.SetupListeners(ctx, pisces.Logger())

	err := crawler.Visit(ctx, server.URL, pisces.Logger())
	require.NoError(t, err)

	visit := crawler.LastVisit()
	assert.NotNil(t, visit)
	assert.NotEmpty(t, visit.RequestedURL)
	assert.NotEmpty(t, visit.Location)
	assert.Empty(t, visit.RedirectLocations)
	assert.NotContains(t, visit.InitialBody, "Hello world!")
	assert.Contains(t, visit.Body, "Hello world!")
	assert.Nil(t, visit.CertificateInfo)

	scriptIdx := slices.IndexFunc(visit.Assets, matchAsset("script.js"))
	scriptAsset := visit.Assets[scriptIdx]
	assert.Equal(t, "Script", scriptAsset.ResourceType)
	assert.Equal(t, int64(200), scriptAsset.ResponseStatus)
	assert.NotEmpty(t, scriptAsset.RequestHeaders)
	assert.NotEmpty(t, scriptAsset.ResponseHeaders)
	assert.NotEmpty(t, scriptAsset.Body)
	assert.NotEmpty(t, scriptAsset.InitiatorURL)
	assert.Nil(t, scriptAsset.CertificateInfo)

	styleIdx := slices.IndexFunc(visit.Assets, matchAsset("style.css"))
	styleAsset := visit.Assets[styleIdx]
	assert.Equal(t, "Stylesheet", styleAsset.ResourceType)
	assert.Equal(t, int64(200), styleAsset.ResponseStatus)
	assert.NotEmpty(t, styleAsset.RequestHeaders)
	assert.NotEmpty(t, styleAsset.ResponseHeaders)
	assert.NotEmpty(t, styleAsset.Body)
	assert.NotEmpty(t, styleAsset.InitiatorURL)
	assert.Nil(t, scriptAsset.CertificateInfo)
}

func TestCrawlerLastVisitWithoutAnyVisits(t *testing.T) {
	t.Parallel()

	crawler := NewCrawler("pisces", 1920, 1080)

	assert.Nil(t, crawler.LastVisit())
}

func TestCrawlerCapturesAllSlowSubresources(t *testing.T) {
	const total = 30

	server := slowSubresourceServer(t, total)

	visit := collectVisit(t, server.URL)

	count := 0

	for _, a := range visit.Assets {
		if !strings.Contains(a.URL, "/img/") {
			continue // e.g. /favicon.ico
		}

		count++

		assert.Equal(t, int64(200), a.ResponseStatus, a.URL)
		assert.NotEmpty(t, a.Body, "asset %s has no body", a.URL)
	}

	assert.Equal(t, total, count)
	assert.NotEmpty(t, visit.InitialBody)
}

func TestCrawlerAssetOrderIsDeterministic(t *testing.T) {
	server := slowSubresourceServer(t, 30)

	urls := func() []string {
		visit := collectVisit(t, server.URL)
		out := make([]string, 0, len(visit.Assets))

		for _, a := range visit.Assets {
			if strings.Contains(a.URL, "/img/") {
				out = append(out, a.URL)
			}
		}

		return out
	}

	first, second := urls(), urls()

	require.Len(t, first, 30)
	assert.Equal(t, first, second)
	assert.True(t, slices.IsSorted(first), "assets are in document order")
}

// collectVisit visits url with a fresh browser and returns the captured visit.
func collectVisit(t *testing.T, url string) *Visit {
	t.Helper()

	pctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ctx, cancelCtx := chromedp.NewContext(pctx)
	defer cancelCtx()

	crawler := NewCrawler("pisces", 1280, 720)
	crawler.SetupListeners(ctx, pisces.Logger())

	require.NoError(t, crawler.Visit(ctx, url, pisces.Logger()))

	visit := crawler.LastVisit()
	require.NotNil(t, visit)

	return visit
}
