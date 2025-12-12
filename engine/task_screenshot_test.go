package engine

import (
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformScreenshotTask(t *testing.T) {
	t.Parallel()

	server := piscestest.NewTestWebServer("simple")
	task := NewTask("screenshot", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	sr, err := performScreenshotTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.NotEmpty(t, sr.Buffer)
}
