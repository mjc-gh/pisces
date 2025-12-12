package browser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidDeviceType(t *testing.T) {
	t.Parallel()

	assert.True(t, IsValidDeviceType("desktop"))
	assert.True(t, IsValidDeviceType("mobile"))
	assert.False(t, IsValidDeviceType("tablet"))
	assert.False(t, IsValidDeviceType(""))
	assert.False(t, IsValidDeviceType("invalid"))
}

func TestIsValidDeviceSize(t *testing.T) {
	t.Parallel()

	assert.True(t, IsValidDeviceSize("large"))
	assert.True(t, IsValidDeviceSize("medium"))
	assert.True(t, IsValidDeviceSize("small"))
	assert.False(t, IsValidDeviceSize(""))
	assert.False(t, IsValidDeviceSize("xlarge"))
	assert.False(t, IsValidDeviceSize("invalid"))
}

func TestDimensionsFromDeviceProfile(t *testing.T) {
	t.Parallel()

	h, w := DimensionsFromDeviceProfile("desktop", "large")
	assert.Equal(t, 1920, h)
	assert.Equal(t, 1080, w)

	h, w = DimensionsFromDeviceProfile("desktop", "medium")
	assert.Equal(t, 1536, h)
	assert.Equal(t, 864, w)

	h, w = DimensionsFromDeviceProfile("desktop", "small")
	assert.Equal(t, 1280, h)
	assert.Equal(t, 720, w)

	h, w = DimensionsFromDeviceProfile("desktop", "")
	assert.Equal(t, 1536, h)
	assert.Equal(t, 864, w)

	h, w = DimensionsFromDeviceProfile("mobile", "large")
	assert.Equal(t, 430, h)
	assert.Equal(t, 932, w)

	h, w = DimensionsFromDeviceProfile("mobile", "medium")
	assert.Equal(t, 390, h)
	assert.Equal(t, 844, w)

	h, w = DimensionsFromDeviceProfile("mobile", "small")
	assert.Equal(t, 375, h)
	assert.Equal(t, 812, w)

	h, w = DimensionsFromDeviceProfile("mobile", "")
	assert.Equal(t, 390, h)
	assert.Equal(t, 844, w)

	h, w = DimensionsFromDeviceProfile("", "medium")
	assert.Equal(t, 1536, h)
	assert.Equal(t, 864, w)

	h, w = DimensionsFromDeviceProfile("", "")
	assert.Equal(t, 1536, h)
	assert.Equal(t, 864, w)

	h, w = DimensionsFromDeviceProfile("invalid", "invalid")
	assert.Equal(t, 1280, h)
	assert.Equal(t, 720, w)
}

func TestUserAgent(t *testing.T) {
	t.Parallel()

	ua := UserAgent("desktop", "chrome")
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36", ua)

	ua = UserAgent("desktop", "")
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36", ua)

	ua = UserAgent("", "chrome")
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36", ua)

	ua = UserAgent("", "")
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36", ua)

	ua = UserAgent("mobile", "chrome")
	assert.Equal(t, "Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Mobile/15E148 Safari/604", ua)

	ua = UserAgent("mobile", "")
	assert.Equal(t, "Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Mobile/15E148 Safari/604", ua)

	ua = UserAgent("desktop", "firefox")
	assert.Empty(t, ua)

	ua = UserAgent("mobile", "safari")
	assert.Empty(t, ua)

	ua = UserAgent("invalid", "chrome")
	assert.Empty(t, ua)
}
