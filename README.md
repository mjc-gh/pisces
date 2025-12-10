# pisces

[![.github/workflows/ci.yml](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml/badge.svg)](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml)
[![GitHub
Release](https://img.shields.io/github/v/release/mjc-gh/pisces?label=Latest%20Release)](https://github.com/mjc-gh/pisces/releases)

A tool for analyzing phishing attack sites. It's built with `chromedp`
and uses the Chrome DevTools Protocol to automate the browser.

## Development

You can build the CLI tool with the following:

```
make build.cli
./build/pisces -h
```

### Dockerfile

The Dockerfile supports running Pisces as a container. The container adds the built binary to a `chromedp` headless container to run the scanner.

Example:
```
docker run --rm \
          -v "$(pwd)":/app \
          pisces:latest \
          analyze https://google.com
```

This also works with the Make command `make run.docker ARGS="[args]"`.

### Dockerfile.dev

`Dockerfile.dev` works along with the `make test.docker` command to run the integration tests within a container.

## Usage

```
NAME:
   pisces - A tool for analyzing phishing sites

USAGE:
   pisces [global options] [command [command options]]

VERSION:
   0.0.3

COMMANDS:
   analyze     Analyze and interact one or more URLs for phishing
   collect     Collect HTML and assets for one or more URLs
   screenshot  Screenshot one or more URLs
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Flags

All commands accept the following flags:

```
   --debug, -d                 (default: false)
   --remote, -r                (default: false)
   --headfull, -H              (default: false)
   --concurrency int, -c int   (default: 0)
   --port int                  (default: 9222)
   --device-type string        (default: "desktop")
   --device-size string        (default: "large")
   --host string               (default: "127.0.0.1")
   --user-agent string         (default: "chrome")
   --help, -h                  show help
```

See command-specific help for output flags.

The `--remote` flag will use a `RemoteAllocator`. It's primarily used
with the [`chromedp` headless shell
container](https://github.com/chromedp/docker-headless-shell) and is
used in conjunction with the `--host` and `--port` flags.

The `--headfull` tag is mutually exclusive with the `--remote` flag.
When used locally, it will enable the Chrome window and you will see the
browser in action.

### Result Output

Below is an example of the `analyze` task's output for a single visit to `login.yahoo.com`. Some of the assets and response details have been truncated for brevity:

```json
{
  "clipboard_texts": [],
  "forms": [
    {
      "action": "https://login.yahoo.com/",
      "method": "post",
      "class": "pure-form",
      "id": "login-username-form",
      "fields": [
        {
          "name": "browser-fp-data",
          "type": "hidden",
          "value": "{\"language\":\"en-US\",\"colorDepth\":24,\"deviceMemory\":8,\"pixelRatio..."
        },
        {
          "name": "crumb",
          "type": "hidden",
          "value": "oS7da9vywmzAAZRwZtT6hA"
        },
        {
          "name": "acrumb",
          "type": "hidden",
          "value": "aQQNkjc6"
        },
        {
          "name": "sessionIndex",
          "type": "hidden",
          "value": "QQ--"
        },
        {
          "name": "deviceCapability",
          "type": "hidden",
          "value": "{\"pa\":{\"status\":true},\"isWebAuthnSupported\":true}"
        },
        {
          "name": "countryCodeIntl",
          "type": "select-one",
          "value": "US"
        },
        {
          "name": "username",
          "type": "text",
          "value": ""
        },
        {
          "name": "passwd",
          "type": "password",
          "value": ""
        },
        {
          "name": "signin",
          "type": "submit",
          "value": "Next"
        },
        {
          "name": "persistent",
          "type": "checkbox",
          "value": "y"
        },
        {
          "name": "",
          "type": "checkbox",
          "value": "on"
        }
      ]
    }
  ],
  "head": {
    "title": "Login - Sign in to Yahoo",
    "description": "Sign in to access the best in class Yahoo Mail, breaking local, ...",
    "favicon_url": "https://s.yimg.com/wm/mbr/images/yahoo-yep-favicon-v1.ico",
    "shortcut_icon_url": "https://s.yimg.com/wm/mbr/images/yahoo-yep-favicon-v1.ico",
    "viewport": "initial-scale=1, maximum-scale=1, user-scalable=0, shrink-to-fit..."
  },
  "links": [
    {
      "href": "https://www.yahoo.com/",
      "text": "\n            \n            \n        "
    },
    {
      "href": "https://help.yahoo.com/kb/index?locale=en_US&page=product&y=PROD...",
      "text": "Help"
    },
    {
      "href": "https://legal.yahoo.com/us/en/yahoo/terms/otos/index.html",
      "text": "Terms",
      "class": "universal-header-links"
    },
    {
      "href": "https://legal.yahoo.com/us/en/yahoo/privacy/index.html",
      "text": "Privacy",
      "class": "universal-header-links privacy-link"
    },
    {
      "href": "https://login.yahoo.com/forgot?done=https%3A%2F%2Fwww.yahoo.com",
      "text": "Forgot username?"
    },
    {
      "href": "https://login.yahoo.com/account/create?specId=yidregsimplified&d...",
      "text": "Create an account",
      "class": "pure-button puree-button-secondary challenge-button"
    }
  ],
  "visible_text": "Help Terms Privacy Sign in using",
  "requested_url": "http://login.yahoo.com",
  "location": "https://login.yahoo.com/",
  "redirect_locations": [
    {
      "status_code": 307,
      "location": "https://login.yahoo.com/"
    }
  ],
  "certificate_info": {
    "protocol": "TLS 1.3",
    "issuers": "DigiCert Global G2 TLS RSA SHA256 2020 CA1",
    "subject_name": "login.yahoo.com",
    "valid_from": "2025-12-02T19:00:00-05:00",
    "valid_to": "2026-03-04T18:59:59-05:00",
    "sans": [
      "login.yahoo.com",
      "*.pr.login.yahoo.com",
      "*.pr.login.aol.com",
      "*.pr.login.engadget.com"
    ]
  },
  "body": "<html id=\"Stencil\" lang=\"en-US\" ",
  "initial_body": "<!DOCTYPE html>\n<html id=\"Stenci",
  "assets": [
    {
      "url": "https://s.yimg.com/ss/analytics3",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "*.yahoo.com",
        "valid_from": "2025-12-02T19:00:00-05:00",
        "valid_to": "2026-01-21T18:59:59-05:00",
        "sans": [
          "*.yahoo.com",
          "*.www.yahoo.com",
          "ymail.com",
          "s.yimg.com"
        ]
      },
      "resource_type": "Script",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "response_headers": {
        "accept-ranges": "bytes",
        "age": "394",
        "ats-carp-promotion": "1, 1",
        "cache-control": "max-age=600",
        "content-encoding": "gzip",
        "content-length": "26369",
        "content-type": "application/javascript",
        "date": "Tue, 09 Dec 2025 03:51:00 GMT",
        "etag": "\"08a736a2a0c457e75316d89e9f2c223b-df\"",
        "last-modified": "Mon, 08 Dec 2025 18:20:09 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin, Accept-Encoding",
        "x-amz-id-2": "vHGh9YQKe/EBI00Ly11fYlGChs3X10XcGP+MntHIv/YLUHbE/SUyD6OXuAM6h2wg...",
        "x-amz-request-id": "CRYR534ZCY72TQM4",
        "x-amz-server-side-encryption": "AES256",
        "x-amz-version-id": "wovds7_PEksTVXYS80umv6C.KYiO6hSd",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "initiator_url": "https://login.yahoo.com/",
      "body": "(function(Re){typeof define==\"fu"
    },
    {
      "url": "https://ups.analytics.yahoo.com/",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert Global G2 TLS RSA SHA256 2020 CA1",
        "subject_name": "*.pubgw.ads.yahoo.com",
        "valid_from": "2025-11-30T19:00:00-05:00",
        "valid_to": "2026-01-21T18:59:59-05:00",
        "sans": [
          "*.pubgw.ads.yahoo.com",
          "client.tools.advertising.yahoo.com",
          "*.fc.yahoo.com",
          "fc.yahoo.com"
        ]
      },
      "resource_type": "Fetch",
      "request_headers": {
        "Accept": "application/json",
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "response_headers": {
        "access-control-allow-credentials": "true",
        "access-control-allow-origin": "https://login.yahoo.com",
        "age": "0",
        "content-type": "application/json",
        "date": "Tue, 09 Dec 2025 03:57:33 GMT",
        "p3p": "CP=NOI DSP COR LAW CURa DEVa TAIa PSAa PSDa OUR BUS UNI COM NAV",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "body": "{\"axid\": \"eS1MbXdzQ2IxRTJ1R2xuZW"
    },
    {
      "url": "https://udc.yahoo.com/v2/public/",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "yahoo.com",
        "valid_from": "2025-10-15T20:00:00-04:00",
        "valid_to": "2026-04-08T19:59:59-04:00",
        "sans": [
          "yahoo.com",
          "tw.mobi.yahoo.com",
          "test.geo.yahoo.com",
          "qa2.my.yahoo.com"
        ]
      },
      "resource_type": "XHR",
      "request_headers": {
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "response_headers": {
        "access-control-allow-credentials": "true",
        "access-control-allow-origin": "https://login.yahoo.com",
        "age": "1",
        "cache-control": "no-store, no-cache, private, max-age=0",
        "date": "Tue, 09 Dec 2025 03:57:32 GMT",
        "expires": "-1",
        "p3p": "policyref=\"http://info.yahoo.com/w3c/p3p.xml\", CP=\"CAO DSP COR C...",
        "pragma": "no-cache",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-envoy-upstream-service-time": "1"
      },
      "response_status": 204
    },
    {
      "url": "https://s.yimg.com/wm/mbr/2df035",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "*.yahoo.com",
        "valid_from": "2025-12-02T19:00:00-05:00",
        "valid_to": "2026-01-21T18:59:59-05:00",
        "sans": [
          "*.yahoo.com",
          "*.www.yahoo.com",
          "ymail.com",
          "s.yimg.com"
        ]
      },
      "resource_type": "Stylesheet",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "response_headers": {
        "accept-ranges": "bytes",
        "age": "628660",
        "ats-carp-promotion": "1, 1",
        "cache-control": "public,max-age=31536000",
        "content-encoding": "gzip",
        "content-length": "140661",
        "content-type": "text/css",
        "date": "Mon, 01 Dec 2025 21:19:54 GMT",
        "etag": "\"c68492486108489b1f71307898c5317d-df\"",
        "last-modified": "Tue, 25 Nov 2025 22:29:15 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Accept-Encoding",
        "x-amz-id-2": "rhznM4nybf3xxv43RV89ApQCzs53JGqSIaXypeNijhN/dskAS5/ndSgiiKIhImJ7...",
        "x-amz-request-id": "8QNRXAHAAKX93P6H",
        "x-amz-server-side-encryption": "AES256",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "initiator_url": "https://login.yahoo.com/",
      "body": "@font-face{font-family:\"Yahoo Sa"
    },
    {
      "url": "https://s.yimg.com/rz/p/yahoo_fr",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "*.yahoo.com",
        "valid_from": "2025-12-02T19:00:00-05:00",
        "valid_to": "2026-01-21T18:59:59-05:00",
        "sans": [
          "*.yahoo.com",
          "*.www.yahoo.com",
          "ymail.com",
          "s.yimg.com"
        ]
      },
      "resource_type": "Image",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "response_headers": {
        "accept-ranges": "bytes",
        "age": "85982",
        "ats-carp-promotion": "1, 1",
        "cache-control": "public,max-age=86400",
        "content-length": "1346",
        "content-type": "image/png",
        "date": "Mon, 08 Dec 2025 04:04:32 GMT",
        "etag": "\"cd166981c96c6d0f4b5a7d798c25878e\"",
        "expires": "Tue, 09 Dec 2025 00:00:00 GMT",
        "last-modified": "Sun, 07 Dec 2025 21:30:30 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-amz-id-2": "NhiQ4UzJ+FQV0JO9TkgP2lAteQrkREJ1ARXw+GZqlQHh3RWSZOplveSj0AojIzsg...",
        "x-amz-request-id": "C26TZ7PP0ZXD0NM1",
        "x-amz-server-side-encryption": "AES256",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "initiator_url": "https://login.yahoo.com/",
      "body": "�PNG\r\n\u001a\n\u0000\u0000\u0000\rIHDR\u0000\u0000\u0000�\u0000\u0000\u0000H\b\u0003\u0000\u0000\u0000�\u0012�"
    },
    {
      "url": "https://s.yimg.com/bw/fonts/yaho",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "*.yahoo.com",
        "valid_from": "2025-12-02T19:00:00-05:00",
        "valid_to": "2026-01-21T18:59:59-05:00",
        "sans": [
          "*.yahoo.com",
          "*.www.yahoo.com",
          "ymail.com",
          "s.yimg.com"
        ]
      },
      "resource_type": "Font",
      "request_headers": {
        "Origin": "https://login.yahoo.com",
        "Referer": "https://s.yimg.com/wm/mbr/2df03510d3a44bebaf3b767b14ccd6be88871c...",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "response_headers": {
        "accept-ranges": "bytes",
        "access-control-allow-origin": "*",
        "age": "192012",
        "ats-carp-promotion": "1, 1",
        "cache-control": "public, max-age=604800",
        "content-length": "34588",
        "content-type": "font/woff2",
        "date": "Sat, 06 Dec 2025 22:37:22 GMT",
        "etag": "\"492a0a160b8da9414134282ef8b62f78\"",
        "last-modified": "Fri, 14 Feb 2025 01:59:26 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-amz-id-2": "+LXUZalbTqY1vJKg6FkYQNQnhSHfQ+ys6L1vqI7McCtQ5jIgLiHX9Xn2EvNDoAy0...",
        "x-amz-request-id": "X59FKGJ27ZSTVRGB",
        "x-amz-server-side-encryption": "AES256",
        "x-amz-version-id": "wUt5QWMHFonqO_Z13wqeP1wud1inlbiC",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "initiator_url": "https://s.yimg.com/wm/mbr/2df035",
      "body": "wOF2\u0000\u0001\u0000\u0000\u0000\u0000�\u001c\u0000\u0011\u0000\u0000\u0000\u0001\u0015\u0014\u0000\u0000��\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000"
    },
    {
      "url": "https://gpt.mail.yahoo.net/sandb",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert Global G2 TLS RSA SHA256 2020 CA1",
        "subject_name": "glp.searchjam.com",
        "valid_from": "2025-12-01T19:00:00-05:00",
        "valid_to": "2026-03-04T18:59:59-05:00",
        "sans": [
          "glp.searchjam.com",
          "external.cap.yahoo.net",
          "external-staging.cap.yahoo.net",
          "qa-glp.searchjam.com"
        ]
      },
      "resource_type": "Document",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "response_headers": {
        "age": "0",
        "content-encoding": "gzip",
        "content-security-policy": "base-uri 'none'; connect-src https:; script-src 'nonce-6gM9Y0q45...",
        "content-type": "text/html; charset=utf-8",
        "date": "Tue, 09 Dec 2025 03:57:33 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Accept-Encoding",
        "x-content-type-options": "nosniff",
        "x-envoy-upstream-service-time": "3",
        "x-omg-env": "norrin-blue--gam-production-bf1-64bfdccc87-chnns",
        "x-xss-protection": "1; mode=block"
      },
      "response_status": 200
    }
  ]
}
```
