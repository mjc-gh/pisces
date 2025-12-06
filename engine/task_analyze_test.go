package engine

import (
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
)

func TestPerformAnalyzeTask(t *testing.T) {
	t.Parallel()

	server := piscestest.NewTestWebServer("simple")
	task := NewTask("analyze", server.URL)

	ar, err := performAnalyzeTask(piscestest.NewTestContext(), &task, pisces.Logger())

	assert.NoError(t, err)
	assert.Equal(t, "A Simple Web Page", ar.Head.Title)
	assert.Equal(t, "Simple Page Hello world!", ar.VisibleText)
}
