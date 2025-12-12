package browser

const SIZE_LARGE = "large"
const SIZE_MEDIUM = "medium"
const SIZE_SMALL = "small"

const PROFILE_DESKTOP = "desktop"
const PROFILE_MOBILE = "mobile"

var ChromeDesktopUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

func IsValidDeviceType(deviceType string) bool {
	return (deviceType == PROFILE_DESKTOP) || (deviceType == PROFILE_MOBILE)
}

func IsValidDeviceSize(deviceSize string) bool {
	return deviceSize == SIZE_LARGE || deviceSize == SIZE_MEDIUM || deviceSize == SIZE_SMALL
}

func DimensionsFromDeviceProfile(deviceType, deviceSize string) (int, int) {
	switch deviceType {
	case PROFILE_DESKTOP, "":
		switch deviceSize {
		case SIZE_LARGE:
			return 1920, 1080
		case SIZE_MEDIUM, "":
			return 1536, 864
		case SIZE_SMALL:
			return 1280, 720
		}
	case PROFILE_MOBILE:
		switch deviceSize {
		case SIZE_LARGE:
			return 430, 932
		case SIZE_MEDIUM, "":
			return 390, 844
		case SIZE_SMALL:
			return 375, 812
		}
	}

	return 1280, 720
}

func UserAgent(deviceType, userAgentAlias string) string {
	switch userAgentAlias {
	case "chrome", "":
		switch deviceType {
		case PROFILE_DESKTOP, "":
			return ChromeDesktopUserAgent
		case PROFILE_MOBILE:
			return "Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Mobile/15E148 Safari/604"
		}
	}

	return ""
}
