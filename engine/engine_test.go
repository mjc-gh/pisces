package engine

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWithRemoteAllocator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		host        string
		port        int
		expectedUrl string
	}{
		{
			name:        "standard host and port",
			host:        "localhost",
			port:        9222,
			expectedUrl: "http://localhost:9222/json/version",
		},
		{
			name:        "IP address host",
			host:        "192.168.1.100",
			port:        8080,
			expectedUrl: "http://192.168.1.100:8080/json/version",
		},
		{
			name:        "custom port",
			host:        "example.com",
			port:        3000,
			expectedUrl: "http://example.com:3000/json/version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New(1, WithRemoteAllocator(tt.host, tt.port))

			assert.Equal(t, tt.expectedUrl, engine.config.remoteURL)
		})
	}
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	t.Run("sets logger on engine", func(t *testing.T) {
		logger := zerolog.Nop()

		engine := New(1, WithLogger(&logger))

		assert.NotNil(t, engine.logger)
		assert.Equal(t, &logger, engine.logger)
	})
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("creates engine with valid concurrency", func(t *testing.T) {
		concurrency := 5
		engine := New(concurrency)

		assert.NotNil(t, engine)
		assert.Equal(t, concurrency, engine.config.concurrency)
		assert.NotNil(t, engine.results)
		assert.NotNil(t, engine.tasks)
	})

	t.Run("sets minimum concurrency to 1 when less than 1", func(t *testing.T) {
		engine := New(0)
		assert.Equal(t, 1, engine.config.concurrency)

		engine = New(-5)
		assert.Equal(t, 1, engine.config.concurrency)
	})

	t.Run("creates engine with multiple options", func(t *testing.T) {
		logger := zerolog.Nop()
		host := "localhost"
		port := 9222
		concurrency := 3

		engine := New(
			concurrency,
			WithRemoteAllocator(host, port),
			WithLogger(&logger),
		)

		assert.Equal(t, concurrency, engine.config.concurrency)
		assert.Equal(t, "http://localhost:9222/json/version", engine.config.remoteURL)
		assert.Equal(t, &logger, engine.logger)
	})

	t.Run("creates engine with no options", func(t *testing.T) {
		engine := New(2)

		assert.NotNil(t, engine)
		assert.Equal(t, 2, engine.config.concurrency)
		assert.Empty(t, engine.config.remoteURL)
		assert.NotNil(t, engine.logger)
	})


	t.Run("channels have correct buffer sizes", func(t *testing.T) {
		concurrency := 10
		engine := New(concurrency)

		// results channel should be unbuffered
		assert.Equal(t, 0, cap(engine.results))

		// tasks channel should be buffered with concurrency size
		assert.Equal(t, concurrency, cap(engine.tasks))
	})
}

func TestOptions_CanBeComposed(t *testing.T) {
	t.Parallel()

	t.Run("options can be created and applied separately", func(t *testing.T) {
		logger := zerolog.Nop()

		opts := []Option{
			WithRemoteAllocator("192.168.1.1", 8080),
			WithLogger(&logger),
		}

		engine := New(4, opts...)

		assert.Equal(t, 4, engine.config.concurrency)
		assert.Equal(t, "http://192.168.1.1:8080/json/version", engine.config.remoteURL)
		assert.Equal(t, &logger, engine.logger)
	})
}

func TestWithRemoteAllocator_UrlFormat(t *testing.T) {
	t.Run("formats URL correctly with various inputs", func(t *testing.T) {
		engine := New(1)

		// Apply option after creation
		opt := WithRemoteAllocator("test.example.com", 9999)
		opt(engine)

		assert.Equal(t, "http://test.example.com:9999/json/version", engine.config.remoteURL)
	})
}
