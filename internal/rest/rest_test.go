package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mjc-gh/pisces"
	"github.com/mjc-gh/pisces/engine"
	"github.com/mjc-gh/pisces/internal/piscestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskRoutes(t *testing.T) {
	server := piscestest.NewTestWebServer("simple")
	defer server.Close()

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	logger := pisces.SetupLogger(false)
	e := engine.New(1, engine.WithLogger(logger))
	e.Start(ctx)
	defer e.Shutdown()

	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{
			name:        "JSON body",
			contentType: "application/json",
			body:        `{"url": "{{URL}}"}`,
		},
		{
			name:        "form URL encoded body",
			contentType: "application/x-www-form-urlencoded",
			body:        "url={{URL_ENCODED}}",
		},
		{
			name:        "plain text body",
			contentType: "text/plain",
			body:        "{{URL}}",
		},
	}

	taskPaths := []string{"/analyze", "/collect"}

	for _, tt := range tests {
		for _, tp := range taskPaths {
			t.Run(tt.name, func(t *testing.T) {
				// Replace URL placeholders with actual server URL
				body := strings.ReplaceAll(tt.body, "{{URL}}", server.URL)
				body = strings.ReplaceAll(body, "{{URL_ENCODED}}", url.QueryEscape(server.URL))

				router := setupRouter("test", e, logger)
				w := httptest.NewRecorder()
				req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, tp, strings.NewReader(body))
				require.NoError(t, err)

				req.Header.Set("Content-Type", tt.contentType)
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)

				response := map[string]any{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				result, ok := response["result"].(map[string]any)
				require.True(t, ok, "result missing from response")

				assets, ok := result["assets"].([]any)
				require.True(t, ok, "assets type conversion failed")
				assert.NotEmpty(t, assets)
			})
		}
	}
}
