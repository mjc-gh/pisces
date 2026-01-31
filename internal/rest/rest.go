package rest

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	zerologgin "github.com/go-mods/zerolog-gin"
	pe "github.com/mjc-gh/pisces/engine"
	"github.com/rs/zerolog"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

const (
	requestServerErr   = "server error"
	requestUrlParamErr = "url parameter missing"
)

func StartServer(version, addr string, engine *pe.Engine, l *zerolog.Logger) error {
	r := setupRouter(version, engine, l)

	return r.Run(addr)
}

func setupRouter(version string, engine *pe.Engine, l *zerolog.Logger) *gin.Engine {
	// Add tags to logger for web
	logger := l.With().
		Str("service", "pisces-web").
		Str("source", "go").
		Logger()

	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()
	r.Use(defaultHeaderMiddleware(version))
	r.Use(gin.Recovery())
	r.Use(zerologgin.LoggerWithOptions(&zerologgin.Options{
		Name:          "server",
		FieldsExclude: []string{"body", "payload", "referer"},
		Logger:        &logger,
	}))

	// TODO
	// gin.SetMode(gin.ReleaseMode)

	// Setup health check route
	r.GET("/health_check", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Define 404 route for root
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found")
	})

	// Setup route for analyze
	r.POST("/analyze", func(c *gin.Context) {
		url, ok := getAndValidateURL(c, &logger)
		if !ok {
			return
		}

		// Create a new task and add it the engine
		t := pe.NewTask("analyze",
			url,
			pe.WithOutChannel(),
			pe.WithParams(parseAnalyzeParams(c)))

		engine.Add(t)

		// Wait for and respond with task result
		if err := respondWithTaskResult(c, t); err != nil {
			logger.Warn().Msgf("analyze error: %v", err)
		}
	})

	// Setup route for collect
	r.POST("/collect", func(c *gin.Context) {
		url, ok := getAndValidateURL(c, &logger)
		if !ok {
			return
		}

		// Create a new task and add it the engine
		t := pe.NewTask("collect",
			url,
			pe.WithOutChannel())

		engine.Add(t)

		// Wait for and respond with task result
		if err := respondWithTaskResult(c, t); err != nil {
			logger.Warn().Msgf("collect error: %v", err)
		}
	})

	// Setup route for screenshot
	r.POST("/screenshot", func(c *gin.Context) {
		url, ok := getAndValidateURL(c, &logger)
		if !ok {
			return
		}

		// Create a new task and add it to the engine
		t := pe.NewTask("screenshot", url, pe.WithOutChannel())
		engine.Add(t)

		result := t.Result()
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{requestServerErr})
			logger.Warn().Msgf("screenshot error: %v", result.Error)

			return
		}

		sr, ok := result.Result.(*pe.ScreenshotResult)
		if !ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{requestServerErr})
			logger.Warn().Msg("screenshot result cast failed")

			return
		}

		// Encode buffer to base64 and wrap in response
		encodedImage := base64.StdEncoding.EncodeToString(*sr.Buffer)
		result.Result = map[string]string{"image": encodedImage}

		// Return the modified result
		c.JSON(http.StatusOK, result)
	})

	return r
}

func getAndValidateURL(c *gin.Context, logger *zerolog.Logger) (string, bool) {
	url, err := extractURL(c)
	if err != nil {
		logger.Warn().Msgf("analyze error: %v", err)

		c.JSON(http.StatusInternalServerError, ErrorResponse{requestServerErr})

		return "", false
	} else if url == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{requestUrlParamErr})

		return "", false
	}

	return url, true
}

func respondWithTaskResult(c *gin.Context, task pe.Task) error {
	// Get result from out channel
	r := task.Result()

	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process request",
		})

		return r.Error
	}

	c.JSON(http.StatusOK, r)

	return nil
}

func defaultHeaderMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Server", fmt.Sprintf("pisces-web (version %s)", version))
	}
}

func parseAnalyzeParams(c *gin.Context) map[string]any {
	params := map[string]any{}

	if wait := c.Query("wait"); wait != "" {
		if val, err := strconv.Atoi(wait); err == nil {
			params["wait"] = val
		}
	}

	if maxFormSubmits := c.Query("max-form-submits"); maxFormSubmits != "" {
		if val, err := strconv.Atoi(maxFormSubmits); err == nil {
			params["max-form-submits"] = val
		}
	}

	if clipboard := c.Query("clipboard"); clipboard != "" {
		if val, err := strconv.ParseBool(clipboard); err == nil {
			params["clipboard"] = val
		}
	}

	if forms := c.Query("forms"); forms != "" {
		if val, err := strconv.ParseBool(forms); err == nil {
			params["forms"] = val
		}
	}

	return params
}

func extractURL(c *gin.Context) (string, error) {
	contentType := c.GetHeader("Content-Type")

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		var data struct {
			URL string `json:"url"`
		}

		if err := c.ShouldBindJSON(&data); err != nil {
			return "", err
		}

		return data.URL, nil

	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		return c.PostForm("url"), nil

	default:
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(body)), nil
	}
}
