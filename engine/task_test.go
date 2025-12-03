package engine

import (
	"context"
	"testing"
	"time"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/stretchr/testify/assert"
)

func TestPerformTaskUnknownType(t *testing.T) {
	t.Parallel()

	task := Task{action: "huh", url: "http://example.com"}
	r := performTask(context.TODO(), &task, pisces.Logger())

	assert.Equal(t, "huh", r.Action)
	assert.Error(t, r.Error)
	assert.NotEmpty(t, r.URL)
	assert.NotEmpty(t, r.Elapsed)
}

func TestTask_SetDevice(t *testing.T) {
	t.Parallel()

	task := Task{
		action:   "test",
		url:      "http://example.com",
		received: time.Now(),
	}

	task.SetDevice("desktop", "large")

	assert.Equal(t, 1920, task.winWidth)
	assert.Equal(t, 1080, task.winHeight)
}

func TestTask_SetUserAgent(t *testing.T) {
	t.Parallel()

	task := Task{
		action:   "test",
		url:      "http://example.com",
		received: time.Now(),
	}

	task.SetUserAgent("desktop", "chrome")

	assert.Equal(t, browser.ChromeDesktopUserAgent, task.userAgent)
}
