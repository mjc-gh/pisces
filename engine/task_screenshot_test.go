package engine

import (
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
)

func TestPerformScreenshotTask(t *testing.T) {
	t.Parallel()

	server := piscestest.NewTestWebServer("simple")
	task := NewTask("screenshot", server.URL)

	sr, err := performScreenshotTask(piscestest.NewTestContext(), &task, pisces.Logger())

	assert.NoError(t, err)
	assert.NotEmpty(t, sr.Buffer)
}
