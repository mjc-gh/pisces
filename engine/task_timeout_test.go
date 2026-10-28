package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hangServer returns a server whose handlers never respond until the test ends
// or the client goes away. Every incoming request is announced on the returned
// channel.
func hangServer(t *testing.T) (*httptest.Server, <-chan struct{}) {
	t.Helper()

	release := make(chan struct{})
	requests := make(chan struct{}, 100)

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		requests <- struct{}{}

		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))

	t.Cleanup(func() {
		close(release)
		server.Close()
	})

	return server, requests
}

func TestTaskTimeout(t *testing.T) {
	server, _ := hangServer(t)
	e := startEngine(t, 3)

	for _, action := range []string{"collect", "analyze", "screenshot"} {
		t.Run(action, func(t *testing.T) {
			task := NewTask(action, server.URL, WithOutChannel(), WithTaskTimeout(500*time.Millisecond))

			start := time.Now()

			e.Add(task)

			res := resultWithin(t, task, 2*time.Second)

			require.Error(t, res.Error)
			require.ErrorIs(t, res.Error, context.DeadlineExceeded)
			assert.Less(t, time.Since(start), 2*time.Second)
		})
	}
}

func TestTaskTimeoutDoesNotAffectFastTasks(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	defer server.Close()

	e := startEngine(t, 1)

	task := NewTask("collect", server.URL, WithOutChannel(), WithTaskTimeout(30*time.Second))
	e.Add(task)

	res := resultWithin(t, task, 20*time.Second)
	require.NoError(t, res.Error)
}

func TestTaskWithoutTimeoutHasNoDeadline(t *testing.T) {
	t.Parallel()

	for _, opts := range [][]TaskOption{nil, {WithTaskTimeout(0)}, {WithTaskTimeout(-time.Second)}} {
		task := NewTask("collect", "example.test", opts...)

		assert.Zero(t, task.timeout)
	}

	assert.Equal(t, time.Second, NewTask("collect", "example.test", WithTaskTimeout(time.Second)).timeout)
}

func TestTaskTimeoutReleasesBrowser(t *testing.T) {
	skipWithoutPS(t)

	server, requests := hangServer(t)

	// Chrome ignores unknown flags, but they show up in its command line, which
	// lets us find the processes that belong to this test.
	marker := "pisces-test-" + uuid.NewString()
	e := startEngine(t, 1, WithExecAllocatorOptions(chromedp.Flag("pisces-test-marker", marker)))

	// Warm up the goroutine count (the workers and the allocator are running).
	baseline := runtime.NumGoroutine()

	// Chrome takes a moment to start, so leave room for it to reach the server.
	task := NewTask("collect", server.URL, WithOutChannel(), WithTaskTimeout(3*time.Second))
	e.Add(task)

	waitForRequest(t, requests)
	assert.Positive(t, chromeProcesses(marker), "expected a browser to be running for the task")

	res := resultWithin(t, task, 10*time.Second)
	require.ErrorIs(t, res.Error, context.DeadlineExceeded)

	assert.Eventually(t, func() bool { return chromeProcesses(marker) == 0 }, 5*time.Second, 50*time.Millisecond,
		"chrome process of the timed out task was not released")
	assert.Eventually(t, func() bool { return runtime.NumGoroutine() <= baseline+2 }, 5*time.Second, 50*time.Millisecond,
		"goroutines did not return to the baseline of %d (now %d)", baseline, runtime.NumGoroutine())
}

// With the local exec allocator every task starts its own Chrome process, so
// the number of browsers is bounded by the engine's concurrency.
func TestEachTaskGetsItsOwnBrowser(t *testing.T) {
	skipWithoutPS(t)

	const tasks = 3

	server, requests := hangServer(t)
	marker := "pisces-test-" + uuid.NewString()
	e := startEngine(t, tasks, WithExecAllocatorOptions(chromedp.Flag("pisces-test-marker", marker)))

	for range tasks {
		e.Add(NewTask("collect", server.URL, WithOutChannel(), WithTaskTimeout(5*time.Second)))
	}

	for range tasks {
		waitForRequest(t, requests) // every task has navigated
	}

	assert.Equal(t, tasks, chromeProcesses(marker))
}

func waitForRequest(t *testing.T, requests <-chan struct{}) {
	t.Helper()

	select {
	case <-requests:
	case <-time.After(15 * time.Second):
		t.Fatal("the browser never requested the page")
	}
}

func skipWithoutPS(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("ps"); err != nil || runtime.GOOS == "windows" {
		t.Skip("ps not available")
	}
}

func TestEngineContextCancelDoesNotBlockShutdown(t *testing.T) {
	for _, tt := range []struct {
		name       string
		readResult bool
	}{
		{"with a results reader", true},
		{"nobody reading results", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server, requests := hangServer(t)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			e := New(1)
			e.Start(ctx)

			// No timeout: only cancelling the engine context can end this task.
			e.Add(NewTask("collect", server.URL))

			waitForRequest(t, requests) // the first task is in flight

			// One more waits in the queue, which holds one task per worker.
			e.Add(NewTask("collect", server.URL+"/queued"))

			var results []Result

			closed := make(chan struct{})

			if tt.readResult {
				go func() {
					defer close(closed)

					for r := range e.Results() {
						results = append(results, r)
					}
				}()
			}

			cancel()

			done := make(chan struct{})
			go func() {
				e.Shutdown()
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("Shutdown did not return within 5s of cancelling the engine context")
			}

			if tt.readResult {
				select {
				case <-closed:
				case <-time.After(5 * time.Second):
					t.Fatal("results channel was not closed")
				}

				for _, r := range results {
					require.Error(t, r.Error)
					require.ErrorIs(t, r.Error, context.Canceled)
				}
			}

			_, open := <-e.Results()
			assert.False(t, open, "results channel is closed")
		})
	}
}

func TestResultErrorMarshalsAsString(t *testing.T) {
	t.Parallel()

	failed := Result{
		Action: "collect",
		URL:    "http://example.test/",
		Error:  fmt.Errorf("navigate: %w", context.DeadlineExceeded),
	}

	for name, marshal := range map[string]func() ([]byte, error){
		"json.Marshal(value)":   func() ([]byte, error) { return json.Marshal(failed) },
		"json.Marshal(pointer)": func() ([]byte, error) { return json.Marshal(&failed) },
		"JSON":                  failed.JSON,
		"PrettyJSON":            failed.PrettyJSON,
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := marshal()
			require.NoError(t, err)

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(raw, &decoded))

			assert.Equal(t, "navigate: context deadline exceeded", decoded["error"])
			assert.Equal(t, "collect", decoded["action"])
			assert.Equal(t, "http://example.test/", decoded["url"])
		})
	}

	ok := Result{Action: "collect", URL: "http://example.test/"}
	raw, err := json.Marshal(ok)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), `"error"`)
}

func TestPerformTaskWrapsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task := NewTask("collect", "example.test")
	res := performTask(ctx, &task, loggerForTest())

	require.Error(t, res.Error)
	assert.ErrorIs(t, res.Error, context.Canceled)
}

// resultWithin waits for the task's result.
func resultWithin(t *testing.T, task Task, d time.Duration) Result {
	t.Helper()

	done := make(chan Result, 1)

	go func() { done <- task.Result() }()

	select {
	case res := <-done:
		return res
	case <-time.After(d):
		t.Fatalf("no result within %s", d)

		return Result{}
	}
}

// chromeProcesses counts the browser (not helper) processes whose command line
// contains marker.
func chromeProcesses(marker string) int {
	out, err := exec.CommandContext(context.Background(), "ps", "-ww", "-axo", "command").Output()
	if err != nil {
		return 0
	}

	n := 0

	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.Contains(line, marker) && !strings.Contains(line, "--type=") {
			n++
		}
	}

	return n
}
