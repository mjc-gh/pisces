package piscestest

import (
	"context"
	"embed"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"

	"github.com/mjc-gh/pisces/internal/browser"
)

//go:embed testdata/*
var testFS embed.FS

type handler struct {
	cookies []http.Cookie
	dir     string
}

type TestWebServerOption func(*handler)

func WithSetCookie(cookie http.Cookie) TestWebServerOption {
	return func(h *handler) {
		h.cookies = append(h.cookies, cookie)
	}
}

func NewTestWebServer(dir string, opts ...TestWebServerOption) *httptest.Server {
	h := handler{make([]http.Cookie, 0), dir}
	for _, opt := range opts {
		opt(&h)
	}

	server := httptest.NewServer(h)

	return server
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// If the client requests "/", serve "index.html" in that directory.
	if path == "/" || path == "" {
		path = "/index.html"
	}

	fullPath := filepath.Join("testdata", h.dir, path)

	file, err := testFS.Open(fullPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)

		return
	} else if _, err := file.Stat(); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "File not found", http.StatusNotFound)

		return
	}

	for _, cookie := range h.cookies {
		http.SetCookie(w, &cookie)
	}

	http.ServeFileFS(w, r, testFS, fullPath)
}

func NewTestContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	remoteUrl, useRemote := os.LookupEnv("PISCES_CHROMEDP_REMOTE_URL")
	if useRemote {
		return browser.StartRemote(ctx, remoteUrl)
	}

	_, useHeadfull := os.LookupEnv("PISCES_HEADFULL")

	return browser.StartLocal(ctx, useHeadfull)
}

func FindByID[T any](id string) func(T) bool {
	return func(item T) bool {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Struct {
			if f := v.FieldByName("ID"); f.IsValid() && f.Kind() == reflect.String {
				return f.String() == id
			}
		}

		return false
	}
}
