package engine

import (
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"
)

// textualMediaTypes are the media types, besides text/* and the +json and +xml
// suffixes, whose bodies are text.
var textualMediaTypes = map[string]bool{
	"application/json":         true,
	"application/javascript":   true,
	"application/x-javascript": true,
	"application/ecmascript":   true,
	"application/xml":          true,
	"application/xhtml+xml":    true,
	"image/svg+xml":            true,
}

// headerValue returns the first value of the named header, ignoring case.
func headerValue(headers map[string]any, name string) string {
	for k, v := range headers {
		if !strings.EqualFold(k, name) {
			continue
		}

		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

// bodyContentType returns the response's Content-Type header, falling back to
// the media type reported by the browser and finally to sniffing the body.
func bodyContentType(headers map[string]any, mimeType string, body []byte) string {
	if ct := strings.TrimSpace(headerValue(headers, "Content-Type")); ct != "" {
		return ct
	}

	if mimeType != "" {
		return mimeType
	}

	if len(body) == 0 {
		return ""
	}

	return http.DetectContentType(body)
}

// isTextual reports whether a body with the given content type can be carried
// in a JSON string: it must have a textual media type and be valid UTF-8.
func isTextual(contentType string, body []byte) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	}

	textual := strings.HasPrefix(mediaType, "text/") ||
		textualMediaTypes[mediaType] ||
		strings.HasSuffix(mediaType, "+json") ||
		strings.HasSuffix(mediaType, "+xml")

	return textual && utf8.Valid(body)
}
