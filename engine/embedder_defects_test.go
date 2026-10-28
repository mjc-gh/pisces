package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
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
	for x := range 16 {
		for y := range 16 {
			img.Set(x, y, color.RGBA{R: uint8(x * 16), G: uint8(y * 16), B: 0xff, A: 0xff})
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
		fmt.Fprint(w, `<!doctype html><html><head><title>binary</title>
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

// TestDefectBinaryBodiesCorruptedByJSON: asset bodies are stored with
// string(bytes), so binary data does not survive a JSON round trip. Fails on
// the original code with a SHA-256 mismatch.
func TestDefectBinaryBodiesCorruptedByJSON(t *testing.T) {
	server, pngBytes, woffBytes := binaryServer(t)
	task := NewTask("collect", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	cr, err := performCollectTask(ctx, &task, pisces.Logger())
	require.NoError(t, err)

	raw, err := json.Marshal(cr)
	require.NoError(t, err)

	var decoded struct {
		Assets []struct {
			URL  string `json:"url"`
			Body string `json:"body"`
		} `json:"assets"`
	}
	require.NoError(t, json.Unmarshal(raw, &decoded))

	want := map[string][]byte{"/image.png": pngBytes, "/font.woff2": woffBytes}
	for path, expected := range want {
		found := false

		for _, a := range decoded.Assets {
			if strings.HasSuffix(a.URL, path) {
				found = true

				assert.Equal(t, sha(expected), sha([]byte(a.Body)), "%s bytes changed after JSON round trip", path)
			}
		}

		assert.True(t, found, "asset %s not captured", path)
	}
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
		fmt.Fprint(w, sb.String())
	})
	mux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		var i int

		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/img/"), "%d.svg", &i)

		if i%3 == 0 {
			time.Sleep(time.Duration(100+(i%5)*100) * time.Millisecond)
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="4" height="4"><title>img %d</title></svg>`, i)
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

// hangServer returns a server whose handlers never respond until the test ends.
func hangServer(t *testing.T) *httptest.Server {
	t.Helper()

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))

	t.Cleanup(func() {
		close(release)
		server.Close()
	})

	return server
}

// TestDefectNoTaskTimeout: a page that never responds blocks a task forever.
// Fails on the original code because no result arrives within the deadline.
func TestDefectNoTaskTimeout(t *testing.T) {
	server := hangServer(t)

	ctx, cancel := piscestest.NewTestContext()
	// Cancelling the context releases the browser (and the stuck worker).
	t.Cleanup(cancel)

	e := New(1)
	e.Start(ctx)

	task := NewTask("collect", server.URL, WithOutChannel())
	e.Add(task)

	done := make(chan Result, 1)
	go func() { done <- task.Result() }()

	select {
	case res := <-done:
		assert.Error(t, res.Error)
	case <-time.After(5 * time.Second):
		t.Fatal("no result within 5s for a page that never responds")
	}
}
