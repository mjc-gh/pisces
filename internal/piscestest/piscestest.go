package piscestest

import (
	"context"
	"embed"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/mjc-gh/pisces/internal/browser"
)

//go:embed testdata/*
var testFS embed.FS

type handler struct {
	dir string
}

func NewTestWebServer(dir string) *httptest.Server {
	server := httptest.NewServer(handler{dir})

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

	http.ServeFileFS(w, r, testFS, fullPath)
}

func NewTestContext() (context.Context, context.CancelFunc) {
	ctx := context.TODO()
	remoteUrl, useRemote := os.LookupEnv("PISCES_CHROMEDP_REMOTE_URL")
	if useRemote {
		return browser.StartRemote(ctx, remoteUrl)
	}

	_, useHeadfull := os.LookupEnv("PISCES_HEADFULL")
	return browser.StartLocal(ctx, useHeadfull)
}
