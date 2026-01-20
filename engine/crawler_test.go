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
