# pisces

[![.github/workflows/ci.yml](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml/badge.svg)](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml)

A tool for analyzing phishing attack sites

# Task Results

### Field Descriptions

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Action | `string` | action | The action performed |
| Elapsed | `time.Duration` | elapsed | Duration in nanoseconds |
| Error | `error` | error | Error message (omitted if nil) |
| URL | `string` | url | The URL accessed |
| Result | `Payload` | result | The task result |

### Notes

- The `Elapsed` field is serialized as an integer representing nanoseconds
- The `Error` field is omitted from JSON when nil due to the `omitempty` tag
- The `Result` field structure depends on the `Payload` type definition

## Analyze Tasks

```
NAME:
   pisces analyze - Analyze one or more URLs

USAGE:
   pisces analyze [options] [url ...]

OPTIONS:
   --debug, -d                 (default: false)
   --remote, -r                (default: false)
   --concurrency int, -c int   (default: 0)
   --port int                  (default: 9222)
   --device-type string        (default: "desktop")
   --device-size string        (default: "large")
   --host string               (default: "127.0.0.1")
   --output string, -o string  (default: "pisces.json")
   --user-agent string         (default: "chrome")
   --help, -h                  show help
```

### Analyze Result Out

```json
{
  "action": "fetch",
  "elapsed": 1500000000,
  "url": "https://api.example.com/data",
  "result": {
    "location": "https://example.com/final",
    "redirectLocations": [
      {
        "status_code": 301,
        "location": "https://example.com/redirect1"
      },
      {
        "status_code": 302,
        "location": "https://example.com/final"
      }
    ],
    "body": "<html><body>Final content</body></html>",
    "bodySize": 38,
    "initialBody": "<html><body>Initial content</body></html>",
    "initialBodySize": 41,
    "assetsCount": 2,
    "assets": {
      "https://example.com/style.css": {
        "url": "https://example.com/style.css",
        "resource_type": "stylesheet",
        "request_headers": {
          "User-Agent": "Mozilla/5.0"
        },
        "response_headers": {
          "Content-Type": "text/css"
        },
        "body": "body { margin: 0; }"
      },
      "https://example.com/script.js": {
        "url": "https://example.com/script.js",
        "resource_type": "script",
        "request_headers": {
          "Accept": "application/javascript"
        },
        "response_headers": {
          "Content-Type": "application/javascript"
        },
        "body": "console.log('hello');"
      }
    }
  }
}
```
