package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformCollectTask(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	task := NewTask("collect", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	cr, err := performCollectTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Equal(t, len(cr.InitialBody), cr.InitialBodyLength)
	assert.Equal(t, len(cr.Body), cr.BodyLength)
	assert.Equal(t, len(cr.Assets), cr.TotalAssets)
}

func TestCollectBinaryAssetsKeepRawBytes(t *testing.T) {
	server, pngBytes, woffBytes := binaryServer(t)
	task := NewTask("collect", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	cr, err := performCollectTask(ctx, &task, pisces.Logger())
	require.NoError(t, err)

	raw, err := json.Marshal(cr)
	require.NoError(t, err)

	var decoded struct {
		Assets []map[string]any `json:"assets"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))

	tests := []struct {
		path        string
		want        []byte
		contentType string
	}{
		{"/image.png", pngBytes, "image/png"},
		{"/font.woff2", woffBytes, "font/woff2"},
	}

	for _, tt := range tests {
		idx := slices.IndexFunc(cr.Assets, matchAsset(tt.path))
		require.GreaterOrEqual(t, idx, 0, "asset %s not captured", tt.path)

		asset := cr.Assets[idx]
		assert.Equal(t, sha(tt.want), sha(asset.RawBody), "%s raw bytes differ from what was served", tt.path)
		assert.Equal(t, tt.contentType, asset.ContentType)
		assert.Equal(t, len(tt.want), asset.Size)
		assert.Empty(t, asset.Body, "binary asset must not populate body")

		// The JSON has content_type and size but no body.
		j := decoded.Assets[idx]
		assert.NotContains(t, j, "body")
		assert.Equal(t, tt.contentType, j["content_type"])
		assert.InDelta(t, float64(len(tt.want)), j["size"], 0)
	}
}

func TestCollectTextAssetsKeepBody(t *testing.T) {
	mux := http.NewServeMux()
	text := map[string]struct{ contentType, body string }{
		"/app.js":    {"text/javascript; charset=utf-8", "document.title = 'caf\u00e9';"},
		"/style.css": {"text/css; charset=utf-8", "body { color: red }"},
		"/data.json": {"application/json", `{"ok":true}`},
		"/logo.svg":  {"image/svg+xml", `<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"></svg>`},
	}

	for path, v := range text {
		mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", v.contentType)
			_, _ = w.Write([]byte(v.body))
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<!doctype html><html><head>
<link rel="stylesheet" href="/style.css">
<link rel="preload" as="fetch" crossorigin href="/data.json">
<script src="/app.js"></script></head><body><img src="/logo.svg"></body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	visit := collectVisit(t, server.URL)

	for path, v := range text {
		idx := slices.IndexFunc(visit.Assets, matchAsset(path))
		require.GreaterOrEqual(t, idx, 0, "asset %s not captured", path)

		asset := visit.Assets[idx]
		assert.Equal(t, v.body, asset.Body, path)
		assert.Equal(t, []byte(v.body), asset.RawBody, path)
		assert.Equal(t, v.contentType, asset.ContentType, path)
		assert.Equal(t, len(v.body), asset.Size, path)
	}

	// The main document is text too.
	assert.Contains(t, visit.InitialBody, "<script src=\"/app.js\">")
	assert.Equal(t, "text/html; charset=utf-8", visit.InitialContentType)
	assert.Equal(t, len(visit.RawInitialBody), visit.InitialSize)
	assert.Equal(t, visit.InitialBody, string(visit.RawInitialBody))
}

func TestCollectBinaryInitialBody(t *testing.T) {
	pngBytes := testPNG(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	}))
	defer server.Close()

	task := NewTask("collect", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	cr, err := performCollectTask(ctx, &task, pisces.Logger())
	require.NoError(t, err)

	assert.Equal(t, sha(pngBytes), sha(cr.RawInitialBody))
	assert.Empty(t, cr.InitialBody)
	assert.Equal(t, "image/png", cr.InitialContentType)
	assert.Equal(t, len(pngBytes), cr.InitialSize)
	assert.Equal(t, len(pngBytes), cr.InitialBodyLength)
}

// A text-only page only gains the additive content_type and size asset keys
// (and initial_content_type / initial_size on the visit); nothing else in the
// JSON changed. TestLegacyJSONShapeUnchanged pins the exact bytes otherwise.
func TestCollectTextOnlyJSONOnlyAddsKeys(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	defer server.Close()

	task := NewTask("collect", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	cr, err := performCollectTask(ctx, &task, pisces.Logger())
	require.NoError(t, err)

	raw, err := json.Marshal(cr)
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(raw, &doc))

	legacyTop := []string{
		"body_length", "initial_body_length", "total_assets", "requested_url", "location",
		"redirect_locations", "certificate_info", "body", "initial_body", "assets",
	}
	legacyAsset := []string{
		"url", "certificate_info", "resource_type", "request_headers", "response_headers",
		"response_status", "initiator_url", "body",
	}

	delete(doc, "initial_content_type")
	delete(doc, "initial_size")

	for k := range doc {
		assert.Contains(t, legacyTop, k)
	}

	assets, ok := doc["assets"].([]any)
	require.True(t, ok)

	require.Len(t, assets, len(cr.Assets))

	for i, a := range assets {
		asset, ok := a.(map[string]any)
		require.True(t, ok)

		// Requests still in flight when the visit ends (e.g. the favicon) have
		// no body, and so no content type or size either.
		if len(cr.Assets[i].RawBody) > 0 {
			assert.Contains(t, asset, "content_type")
			assert.Contains(t, asset, "size")
		}

		delete(asset, "content_type")
		delete(asset, "size")

		for k := range asset {
			assert.Contains(t, legacyAsset, k)
		}
	}
}
