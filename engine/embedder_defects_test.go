package engine

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file reproduces the defects that make the engine unsafe to embed in a
// long running service. Each test fails on the code as it was before the
// embedder-support changes (see the doc comment on each test for the reason).

// testPNG returns a small deterministic PNG image.
func testPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for x := range uint8(16) {
		for y := range uint8(16) {
			img.Set(int(x), int(y), color.RGBA{R: x * 16, G: y * 16, B: 0xff, A: 0xff})
		}
	}

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	return buf.Bytes()
}

// testWOFF2 returns deterministic bytes that start with the WOFF2 magic and
// are followed by pseudo-random data (so they are not valid UTF-8).
func testWOFF2() []byte {
	data := make([]byte, 2048)
	copy(data, "wOF2")

	//nolint:gosec // deterministic test data
	rng := rand.New(rand.NewSource(42))
	for i := 4; i < len(data); i++ {
		data[i] = byte(0x80 | rng.Intn(0x80))
	}

	return data
}

func sha(b []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

// binaryServer serves a page that pulls in a PNG and a WOFF2 file.
func binaryServer(t *testing.T) (*httptest.Server, []byte, []byte) {
	t.Helper()

	pngBytes := testPNG(t)
	woffBytes := testWOFF2()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<!doctype html><html><head><title>binary</title>
<link rel="preload" as="font" type="font/woff2" crossorigin href="/font.woff2">
</head><body><img src="/image.png"></body></html>`)
	})
	mux.HandleFunc("/image.png", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytes)
	})
	mux.HandleFunc("/font.woff2", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "font/woff2")
		_, _ = w.Write(woffBytes)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, pngBytes, woffBytes
}

// slowSubresourceServer serves a page with n images; every third one is slow.
func slowSubresourceServer(t *testing.T, n int) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		var sb strings.Builder

		sb.WriteString("<!doctype html><html><body>")

		for i := range n {
			fmt.Fprintf(&sb, `<img src="/img/%02d.svg">`, i)
		}

		sb.WriteString("</body></html>")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, sb.String())
	})
	mux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		var i int

		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/img/"), "%d.svg", &i)

		if i%3 == 0 {
			time.Sleep(time.Duration(100+(i%5)*100) * time.Millisecond)
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="4" height="4"><title>img %d</title></svg>`, i)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

// TestDefectBodyFetchRace: asset bodies are fetched in goroutines that write to
// the Asset after captureVisit has returned. Fails on the original code under
// `go test -race` with a data race report. Run it with -count=20 -race.
func TestDefectBodyFetchRace(t *testing.T) {
	server := slowSubresourceServer(t, 12)
	task := NewTask("collect", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	cr, err := performCollectTask(ctx, &task, pisces.Logger())
	require.NoError(t, err)

	images := 0

	for _, a := range cr.Assets {
		if !strings.Contains(a.URL, "/img/") {
			continue // e.g. /favicon.ico
		}

		images++

		assert.NotEmpty(t, a.Body, "asset %s has no body", a.URL)
	}

	assert.Equal(t, 12, images)
}
