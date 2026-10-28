package engine

import (
	"bytes"
	"sync"

	"github.com/rs/zerolog"
)

// lockedBuffer is a bytes.Buffer that is safe for concurrent writers.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func loggerForTest() *zerolog.Logger {
	l := zerolog.Nop()

	return &l
}
