package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/rs/zerolog"
)

type Task struct {
	id        uuid.UUID
	action    string
	params    map[string]any
	outChan   chan Result
	url       string
	userAgent string
	winHeight int
	winWidth  int
	received  time.Time
}

type TaskOption func(*Task)

func WithParams(m map[string]any) TaskOption {
	return func(t *Task) {
		maps.Copy(t.params, m)
	}
}

func WithDeviceProperties(deviceType, deviceSize string) TaskOption {
	return func(t *Task) {
		t.winWidth, t.winHeight = browser.DimensionsFromDeviceProfile(deviceType, deviceSize)
	}
}

func WithUserAgent(deviceType, userAgentAlias string) TaskOption {
	return func(t *Task) {
		t.userAgent = browser.UserAgent(deviceType, userAgentAlias)
	}
}

func WithOutChannel() TaskOption {
	return func(t *Task) {
		t.outChan = make(chan Result, 1)
	}
}

func NewTask(action, input string, opts ...TaskOption) Task {
	url := input

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	t := Task{
		id:        uuid.New(),
		action:    action,
		params:    map[string]any{},
		received:  time.Now(),
		url:       url,
		winHeight: 720,
		winWidth:  1280,
	}

	for _, opt := range opts {
		opt(&t)
	}

	return t
}

func (t Task) ID() string {
	return t.id.String()
}

// IntParam will get a task parameter with the given key as an int value.
func (t Task) IntParam(key string, defaultVal int) int {
	val, ok := t.params[key]
	if !ok {
		return defaultVal
	}

	n, ok := val.(int)
	if !ok {
		return defaultVal
	}

	return n
}

func (t Task) BoolParam(key string, defaultVal bool) bool {
	val, ok := t.params[key]
	if !ok {
		return defaultVal
	}

	n, ok := val.(bool)
	if !ok {
		return defaultVal
	}

	return n
}

func (t Task) Result() chan Result {
	return t.outChan
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

type Payload any

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
		return newErrorResult(task, fmt.Errorf("%w: %s", ErrUnknownAction, task.action))
	}

	result.Elapsed = time.Since(task.received)

	return result
}
