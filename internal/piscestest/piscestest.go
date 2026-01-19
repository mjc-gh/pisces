package piscestest

import (
	"context"
	"embed"
	"errors"
	"log"
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
	cookies   []http.Cookie
	redirects map[string]string
	dir       string
}

type TestWebServerOption func(*handler)

func WithSetCookie(cookie http.Cookie) TestWebServerOption {
	return func(h *handler) {
		h.cookies = append(h.cookies, cookie)
	}
}

func WithRedirectFromPOST(path, dest string) TestWebServerOption {
	return func(h *handler) {
		h.redirects[path] = dest
	}
}

func NewTestWebServer(dir string, opts ...TestWebServerOption) *httptest.Server {
	h := handler{
		make([]http.Cookie, 0),
		make(map[string]string),
		dir,
	}

	for _, opt := range opts {
		opt(&h)
	}

	server := httptest.NewServer(h)

	return server
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("piscestest request: %s %s\n", r.Method, r.URL)

	path := r.URL.Path

	// Check for redirects
	if dest, ok := h.redirects[path]; ok {
		http.Redirect(w, r, dest, http.StatusFound)

		return
	}

	// If the client requests "/", serve "index.html" in that directory.
	ext := filepath.Ext(path)
	if path == "/" || path == "" {
		path = "/index.html"
	} else if ext == "" {
		path += ".html"
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
