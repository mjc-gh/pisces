package engine

import (
	"slices"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
)

func matchAsset(fileName string) func(*Asset) bool {
	return func(a *Asset) bool {
		return strings.Contains(a.URL, fileName)
	}
}

func TestCrawlerVisit(t *testing.T) {
	ctx, _ := chromedp.NewContext(piscestest.NewTestContext())
	server := piscestest.NewTestWebServer("simple")
	crawler := NewCrawler("pisces", 1920, 1080)

	err := crawler.Visit(ctx, server.URL, pisces.Logger())
	assert.NoError(t, err)

	visit := crawler.LastVisit()
	assert.NotNil(t, visit)
	assert.NotEmpty(t, visit.RequestedUrl)
	assert.NotEmpty(t, visit.Location)
	assert.Empty(t, visit.RedirectLocations)
	assert.NotContains(t, visit.InitialBody, "Hello world!")
	assert.Contains(t, visit.Body, "Hello world!")

	scriptIdx := slices.IndexFunc(visit.Assets, matchAsset("script.js"))
	scriptAsset := visit.Assets[scriptIdx]
	assert.Equal(t, "Script", scriptAsset.ResourceType)
	assert.NotEmpty(t, scriptAsset.RequestHeaders)
	assert.NotEmpty(t, scriptAsset.ResponseHeaders)
	assert.NotEmpty(t, scriptAsset.Body)
	assert.NotEmpty(t, scriptAsset.InitiatorURL)

	styleIdx := slices.IndexFunc(visit.Assets, matchAsset("style.css"))
	styleAsset := visit.Assets[styleIdx]
	assert.Equal(t, "Stylesheet", styleAsset.ResourceType)
	assert.NotEmpty(t, styleAsset.RequestHeaders)
	assert.NotEmpty(t, styleAsset.ResponseHeaders)
	assert.NotEmpty(t, styleAsset.Body)
	assert.NotEmpty(t, styleAsset.InitiatorURL)
}

func TestCrawlerLastVisitWithoutAnyVisits(t *testing.T) {
	crawler := NewCrawler("pisces", 1920, 1080)

	assert.Nil(t, crawler.LastVisit())
}
