package engine

import (
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformAnalyzeTask(t *testing.T) {
	t.Parallel()

	server := piscestest.NewTestWebServer("simple")
	task := NewTask("analyze", server.URL)
	task.params = map[string]any{"wait": 100}

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ar, err := performAnalyzeTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Equal(t, "A Simple Web Page", ar.Head.Title)
	assert.Equal(t, "Simple Page Hello world!", ar.VisibleText)
}

func TestPerformAnalyzeTaskWithClipboardInteractions(t *testing.T) {
	t.Parallel()

	server := piscestest.NewTestWebServer("fakecaptcha")
	task := NewTask("analyze", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ar, err := performAnalyzeTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Equal(t, []string{
		"msiexec /i https://totally.legit/captcha",
	}, ar.ClipboardTexts)
}
