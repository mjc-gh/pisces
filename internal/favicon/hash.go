package favicon

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/bits"
	"net/http"
	"regexp"

	"github.com/chromedp/chromedp"
)

// MurmurHash3 implements the 32-bit MurmurHash3 algorithm: https://en.wikipedia.org/wiki/MurmurHash
func MurmurHash3(data []byte) int32 {
	const (
		c1 uint32 = 0xcc9e2d51
		c2 uint32 = 0x1b873593
		r1 uint32 = 15
		r2 uint32 = 13
		m  uint32 = 5
		n  uint32 = 0xe6546b64
	)

	var h uint32 = 0 // seed = 0
	length := len(data)
	nblocks := length / 4

	// Process 4-byte blocks
	for i := 0; i < nblocks; i++ {
		k := uint32(data[i*4]) |
			uint32(data[i*4+1])<<8 |
			uint32(data[i*4+2])<<16 |
			uint32(data[i*4+3])<<24

		k *= c1
		k = bits.RotateLeft32(k, int(r1))
		k *= c2

		h ^= k
		h = bits.RotateLeft32(h, int(r2))
		h = h*m + n
	}

	// Process remaining bytes
	tail := data[nblocks*4:]
	var k uint32

	switch len(tail) {
	case 3:
		k ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k ^= uint32(tail[0])
		k *= c1
		k = bits.RotateLeft32(k, int(r1))
		k *= c2
		h ^= k
	}

	// Finalization
	h ^= uint32(length)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16

	// Return as signed int32 for Shodan compatibility
	return int32(h)
}

// addNewlines adds a newline every n characters to match Shodan's base64 format
func addNewlines(s string, n int) string {
	if n <= 0 {
		return s
	}

	re := regexp.MustCompile(fmt.Sprintf("(.{%d})", n))
	result := re.ReplaceAllString(s, "$1\n")

	return result
}

// CalculateHash fetches a favicon from the given URL and calculates its Shodan-compatible hash
func CalculateHash(faviconURL string) (string, error) {
	if faviconURL == "" {
		return "", fmt.Errorf("empty favicon URL")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * 1000000000, // 10 seconds
	}

	resp, err := client.Get(faviconURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch favicon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("favicon fetch returned status %d", resp.StatusCode)
	}

	// Read favicon data
	faviconData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read favicon data: %w", err)
	}

	// Calculate hash using Shodan's method:
	// 1. Base64 encode the favicon
	b64 := base64.StdEncoding.EncodeToString(faviconData)

	// 2. Add newlines every 76 characters (as per Shodan's method)
	b64WithNewlines := addNewlines(b64, 76)

	// 3. Calculate mmh3 hash
	hash := MurmurHash3([]byte(b64WithNewlines))

	return fmt.Sprintf("%d", hash), nil
}

// ExtractFaviconURL extracts the favicon URL from the current page
func ExtractFaviconURL(ctx context.Context, pageURL string) (string, error) {
	var faviconURL string

	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				// Look for link tags with rel="icon" or rel="shortcut icon"
				let link = document.querySelector('link[rel~="icon"]');
				if (link && link.href) {
					return link.href;
				}
				
				// Fallback to default /favicon.ico location
				try {
					const url = new URL(window.location.href);
					return url.origin + '/favicon.ico';
				} catch (e) {
					return '';
				}
			})()
		`, &faviconURL),
	)

	if err != nil {
		return "", err
	}

	return faviconURL, nil
}
