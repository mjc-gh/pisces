package engine

import (
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformCollectTask(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	task := NewTask("collect", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	cr, err := performCollectTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Equal(t, len(cr.InitialBody), cr.InitialBodyLength)
	assert.Equal(t, len(cr.Body), cr.BodyLength)
	assert.Equal(t, len(cr.Assets), cr.TotalAssets)
}
