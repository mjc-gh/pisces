package engine

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformScreenshotTask(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	task := NewTask("screenshot", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	sr, err := performScreenshotTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, sr.Buffer)
}

func TestPerformScreenshotTaskFullPage(t *testing.T) {
	const pageHeight = 3000

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!doctype html><html><body style="margin:0"><div style="height:%dpx;background:#369"></div></body></html>`, pageHeight)
	}))
	defer server.Close()

	pngHeight := func(params map[string]any) int {
		task := NewTask("screenshot", server.URL, WithDeviceProperties("desktop", "small"), WithParams(params))

		ctx, cancel := piscestest.NewTestContext()
		defer cancel()

		sr, err := performScreenshotTask(ctx, &task, pisces.Logger())
		require.NoError(t, err)

		cfg, err := png.DecodeConfig(bytes.NewReader(*sr.Buffer))
		require.NoError(t, err)

		return cfg.Height
	}

	_, viewportHeight := browser.DimensionsFromDeviceProfile("desktop", "small")

	assert.Equal(t, viewportHeight, pngHeight(nil), "default captures the viewport only")
	assert.Equal(t, viewportHeight, pngHeight(map[string]any{"full_page": false}))
	assert.GreaterOrEqual(t, pngHeight(map[string]any{"full_page": true}), pageHeight)
}
