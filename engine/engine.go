package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/rs/zerolog"
)

type config struct {
	concurreny int
	remoteUrl  string
}

type Engine struct {
	browserCancel context.CancelFunc
	config        config
	logger        *zerolog.Logger
	results       chan Result
	tasks         chan Task
	wg            sync.WaitGroup
}

type Option func(*Engine)

func WithRemoteAllocator(host string, port int) Option {
	return func(e *Engine) {
		e.config.remoteUrl = fmt.Sprintf("http://%s:%d/json/version", host, port)
	}
}

func WithLogger(l *zerolog.Logger) Option {
	return func(e *Engine) {
		e.logger = l
	}
}

// Start our engine and goroutines worker pool
func New(concurreny int, opts ...Option) *Engine {
	if concurreny < 1 {
		concurreny = 1
	}

	e := Engine{
		config:  config{concurreny: concurreny},
		results: make(chan Result),
		tasks:   make(chan Task, concurreny),
		wg:      sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(&e)
	}

	return &e
}

func (e *Engine) Start(ctx context.Context) {
	if e.config.remoteUrl != "" {
		ctx, e.browserCancel = browser.StartRemote(ctx, e.config.remoteUrl)
	} else {
		ctx, e.browserCancel = browser.StartLocal(ctx)
	}

	for i := 0; i < e.config.concurreny; i++ {
		e.wg.Add(1)

		go func(idx int, tasks <-chan Task, results chan<- Result, done func(), logger *zerolog.Logger) {
			logger.Debug().Msgf("task worker #%d started", idx)
			defer done()
			defer logger.Debug().Msgf("task worker #%d stopped", idx)

			for task := range tasks {
				logger.Debug().Msgf("task worker #%d got task", idx)
				results <- performTask(ctx, &task, logger)
				logger.Debug().Msgf("task worker #%d sent result", idx)
			}
		}(i+1, e.tasks, e.results, e.wg.Done, e.logger)
	}
}

// Blocks until workers goroutines have completed their tasks
func (e *Engine) Shutdown() {
	e.logger.Debug().Msg("shutdown called")
	defer e.logger.Debug().Msg("shutdown done")

	if e.browserCancel != nil {
		defer e.browserCancel()
	}

	close(e.tasks)
	e.wg.Wait()
	close(e.results)

}

// Get the Results channel
func (e *Engine) Results() chan Result {
	return e.results
}

// Add work to the tasks queue
func (e *Engine) Add(t Task) {
	e.tasks <- t
}
