package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/rs/zerolog"
)

type Task struct {
	action    string
	params    map[string]any
	url       string
	userAgent string
	winHeight int
	winWidth  int
	received  time.Time
}

func NewTask(action, input string) Task {
	url := input

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = fmt.Sprintf("http://%s", url)
	}

	return Task{
		action:    action,
		received:  time.Now(),
		url:       url,
		winHeight: 720,
		winWidth:  1280,
	}
}

func (t *Task) SetDevice(deviceType, deviceSize string) {
	t.winWidth, t.winHeight = browser.DimensionsFromDeviceProfile(deviceType, deviceSize)
}

func (t *Task) SetUserAgent(deviceType, userAgentAlias string) {
	t.userAgent = browser.UserAgent(deviceType, userAgentAlias)
}

// Parsed from task parameters using the "wait" key. Expected to be an int64
// value in milliseconds. Tasking will use this parameter as a timeout value
// when querying chromedp for nodes. When no value is provided, 50 milliseconds
// will be used.
func (t Task) WaitFor() int64 {
	val, ok := t.params["wait"]
	if !ok {
		return 50
	}

	wait, ok := val.(int64)
	if !ok {
		return 50
	}

	return wait
}

type Result struct {
	Action  string        `json:"action"`
	Elapsed time.Duration `json:"elapsed"`
	Error   error         `json:"error,omitempty"`
	URL     string        `json:"url"`
	Result  Payload       `json:"result"`
}

func newErrorResult(task *Task, err error) Result {
	return Result{
		task.action,
		time.Since(task.received),
		err,
		task.url,
		nil,
	}
}

func (r *Result) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Result) PrettyJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "\t")
}

type Payload interface{}

func performTask(ctx context.Context, task *Task, logger *zerolog.Logger) Result {
	logger.Debug().Msgf("perform task: %+v", task)
	defer logger.Debug().Msgf("performed task: %+v", task)

	result := Result{
		Action: task.action,
		URL:    task.url,
	}

	tlog := logger.With().Str("action", task.action).Logger()

	switch task.action {
	case "analyze":
		payload, err := performAnalyzeTask(ctx, task, &tlog)
		if err != nil {
			return newErrorResult(task, err)
		}

		result.Result = &payload

	case "collect":
		payload, err := performCollectTask(ctx, task, &tlog)
		if err != nil {
			return newErrorResult(task, err)
		}

		result.Result = &payload

	case "screenshot":
		payload, err := performScreenshotTask(ctx, task, &tlog)
		if err != nil {
			return newErrorResult(task, err)
		}

		result.Result = &payload

	default:
		// Unknown task -- shouldn't happen in practice
		return newErrorResult(task, fmt.Errorf("unknown action: %s", task.action))
	}

	result.Elapsed = time.Since(task.received)

	return result
}
