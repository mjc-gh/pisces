package engine

import (
	"context"
	"testing"
	"time"

	"github.com/mjc-gh/pisces"
	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		input          string
		opts           []TaskOption
		expectedAction string
		expectedURL    string
		expectedParams map[string]any
		expectedWidth  int
		expectedHeight int
		checkUserAgent bool
		checkReceived  bool
	}{
		{
			name:           "basic task with http prefix",
			action:         "navigate",
			input:          "http://example.com",
			opts:           nil,
			expectedAction: "navigate",
			expectedURL:    "http://example.com",
			expectedParams: map[string]any{},
			expectedWidth:  1280,
			expectedHeight: 720,
			checkReceived:  true,
		},
		{
			name:           "basic task with https prefix",
			action:         "navigate",
			input:          "https://example.com",
			opts:           nil,
			expectedAction: "navigate",
			expectedURL:    "https://example.com",
			expectedParams: map[string]any{},
			expectedWidth:  1280,
			expectedHeight: 720,
			checkReceived:  true,
		},
		{
			name:           "task without http prefix",
			action:         "navigate",
			input:          "example.com",
			opts:           nil,
			expectedAction: "navigate",
			expectedURL:    "http://example.com",
			expectedParams: map[string]any{},
			expectedWidth:  1280,
			expectedHeight: 720,
			checkReceived:  true,
		},
		{
			name:   "task with params",
			action: "navigate",
			input:  "http://example.com",
			opts: []TaskOption{
				WithParams(map[string]any{"timeout": 30, "retry": true}),
			},
			expectedAction: "navigate",
			expectedURL:    "http://example.com",
			expectedParams: map[string]any{"timeout": 30, "retry": true},
			expectedWidth:  1280,
			expectedHeight: 720,
			checkReceived:  true,
		},
		{
			name:   "task with multiple params calls",
			action: "navigate",
			input:  "http://example.com",
			opts: []TaskOption{
				WithParams(map[string]any{"timeout": 30}),
				WithParams(map[string]any{"retry": true}),
			},
			expectedAction: "navigate",
			expectedURL:    "http://example.com",
			expectedParams: map[string]any{"timeout": 30, "retry": true},
			expectedWidth:  1280,
			expectedHeight: 720,
			checkReceived:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			task := NewTask(tt.action, tt.input, tt.opts...)
			after := time.Now()

			assert.Equal(t, tt.expectedAction, task.action)
			assert.Equal(t, tt.expectedURL, task.url)
			assert.Equal(t, tt.expectedParams, task.params)
			assert.Equal(t, tt.expectedWidth, task.winWidth)
			assert.Equal(t, tt.expectedHeight, task.winHeight)

			if tt.checkReceived {
				assert.True(t, task.received.After(before) || task.received.Equal(before))
				assert.True(t, task.received.Before(after) || task.received.Equal(after))
			}
		})
	}
}

func TestTaskIntParam(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		key         string
		defaultVal  int
		expectedVal int
	}{
		{
			name:        "existing int parameter",
			params:      map[string]any{"timeout": 30},
			key:         "timeout",
			defaultVal:  10,
			expectedVal: 30,
		},
		{
			name:        "missing parameter returns default",
			params:      map[string]any{},
			key:         "timeout",
			defaultVal:  10,
			expectedVal: 10,
		},
		{
			name:        "non-int parameter returns default",
			params:      map[string]any{"timeout": "30"},
			key:         "timeout",
			defaultVal:  10,
			expectedVal: 10,
		},
		{
			name:        "zero value int parameter",
			params:      map[string]any{"timeout": 0},
			key:         "timeout",
			defaultVal:  10,
			expectedVal: 0,
		},
		{
			name:        "negative int parameter",
			params:      map[string]any{"timeout": -5},
			key:         "timeout",
			defaultVal:  10,
			expectedVal: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("test", "http://example.com", WithParams(tt.params))
			result := task.IntParam(tt.key, tt.defaultVal)
			assert.Equal(t, tt.expectedVal, result)
		})
	}
}

func TestPerformTaskUnknownType(t *testing.T) {
	t.Parallel()

	task := Task{action: "huh", url: "http://example.com"}
	r := performTask(context.TODO(), &task, pisces.Logger())

	assert.Equal(t, "huh", r.Action)
	assert.Error(t, r.Error)
	assert.NotEmpty(t, r.URL)
	assert.NotEmpty(t, r.Elapsed)
}
