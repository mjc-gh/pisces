package engine

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/mjc-gh/pisces/internal/browser"
	"github.com/rs/zerolog"
)

type config struct {
	concurrency int
	remoteURL   string
	headfull    bool
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
		e.config.remoteURL = fmt.Sprintf("http://%s:%d/json/version", host, port)
	}
}

func WithHeadfullLocalAllocator() Option {
	return func(e *Engine) {
		e.config.headfull = true
	}
}

func WithLogger(l *zerolog.Logger) Option {
	return func(e *Engine) {
		e.logger = l
	}
}

func New(concurrency int, opts ...Option) *Engine {
	if concurrency < 1 {
		concurrency = 1
	}

	e := Engine{
		config: config{
			concurrency: concurrency,
		},
		results: make(chan Result),
		tasks:   make(chan Task, concurrency),
		wg:      sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(&e)
	}

	if e.logger == nil {
		l := zerolog.New(os.Stderr).With().Timestamp().Logger()
		e.logger = &l
	}

	return &e
}

func (e *Engine) Start(ctx context.Context) {
	if e.config.remoteURL != "" {
		ctx, e.browserCancel = browser.StartRemote(ctx, e.config.remoteURL)
	} else {
		ctx, e.browserCancel = browser.StartLocal(ctx, e.config.headfull)
	}

	for i := 0; i < e.config.concurrency; i++ {
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

func (e *Engine) Results() <-chan Result {
	return e.results
}

func (e *Engine) Add(t Task) {
	e.tasks <- t
}
