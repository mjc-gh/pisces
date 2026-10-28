package engine

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chromeExecPath finds a Chrome binary for tests that need to pass an explicit
// ExecPath. It skips the test when none is available.
func chromeExecPath(t *testing.T) string {
	t.Helper()

	if p := os.Getenv("CHROME_PATH"); p != "" {
		return p
	}

	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome", "headless-shell"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	if runtime.GOOS == "darwin" {
		const app = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		if _, err := os.Stat(app); err == nil {
			return app
		}
	}

	t.Skip("no local Chrome binary found; set CHROME_PATH")

	return ""
}

// startEngine starts an engine and registers cleanup that cancels its context
// and shuts it down.
func startEngine(t *testing.T, concurrency int, opts ...Option) *Engine {
	t.Helper()

	nop := zerolog.Nop()
	e := New(concurrency, append([]Option{WithLogger(&nop)}, opts...)...)
	ctx, cancel := context.WithCancel(context.Background())

	e.Start(ctx)

	t.Cleanup(func() {
		cancel()
		e.Shutdown()
	})

	return e
}

func TestWithExecAllocatorOptionsCollectsPage(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	defer server.Close()

	dataDir := t.TempDir()
	e := startEngine(t, 1, WithExecAllocatorOptions(
		chromedp.NoSandbox,
		chromedp.ExecPath(chromeExecPath(t)),
		chromedp.UserDataDir(dataDir),
	))

	task := NewTask("collect", server.URL, WithOutChannel())
	e.Add(task)

	res := task.Result()
	require.NoError(t, res.Error)

	cr, ok := res.Result.(*CollectResult)
	require.True(t, ok)
	assert.Contains(t, cr.Body, "Hello world!")

	// Chrome writes its profile into the directory it was given.
	entries, err := os.ReadDir(dataDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "user data dir was not populated")

	_, err = os.Stat(filepath.Join(dataDir, "Default"))
	assert.NoError(t, err)
}

func TestWithExecAllocatorOptionsBadExecPath(t *testing.T) {
	e := startEngine(t, 1, WithExecAllocatorOptions(
		chromedp.ExecPath(filepath.Join(t.TempDir(), "no-such-chrome")),
	))

	task := NewTask("collect", "http://127.0.0.1:1", WithOutChannel())
	e.Add(task)

	done := make(chan Result, 1)
	go func() { done <- task.Result() }()

	select {
	case res := <-done:
		require.Error(t, res.Error)
		assert.Contains(t, res.Error.Error(), "no-such-chrome")
	case <-time.After(10 * time.Second):
		t.Fatal("task with a bad exec path did not return")
	}
}

func TestWithExecAllocatorOptionsAccumulates(t *testing.T) {
	t.Parallel()

	e := New(1,
		WithExecAllocatorOptions(chromedp.NoSandbox),
		WithExecAllocatorOptions(chromedp.Flag("disable-dev-shm-usage", true), chromedp.DisableGPU),
	)

	assert.Len(t, e.config.execOpts, 3)
	assert.Empty(t, New(1).config.execOpts)
}

func TestWithExecAllocatorOptionsIgnoredWithRemoteAllocator(t *testing.T) {
	t.Parallel()

	var buf lockedBuffer

	logger := zerolog.New(&buf)
	e := New(1,
		WithLogger(&logger),
		WithRemoteAllocator("127.0.0.1", 1),
		WithExecAllocatorOptions(chromedp.NoSandbox),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	e.Start(ctx)
	e.Shutdown()

	assert.Contains(t, buf.String(), "ignored when using a remote allocator")
}
