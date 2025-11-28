# pisces

[![.github/workflows/ci.yml](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml/badge.svg)](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml)
![GitHub Release](https://img.shields.io/github/v/release/mjc-gh/pisces)

A tool for analyzing phishing attack sites

## Development

You can build the CLI tool with the following:

```
make build.cli
./build/pisces -h
```

## Usage

```
NAME:
   pisces - A tool for analyzing phishing sites

USAGE:
   pisces [global options] [command [command options]]

VERSION:
   0.0.0

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
   --concurrency int, -c int   (default: 0)
   --port int                  (default: 9222)
   --device-type string        (default: "desktop")
   --device-size string        (default: "large")
   --host string               (default: "127.0.0.1")
   --user-agent string         (default: "chrome")
```

See command specific help for output flags.

### Result Output

```json
{
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
          "value": "7R0rwaAoHkrnKe/zXO8I1A"
        },
        {
          "name": "acrumb",
          "type": "hidden",
          "value": "VmquGmZi"
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
  "head": {
    "title": "Login - Sign in to Yahoo",
    "description": "Sign in to access the best in class Yahoo Mail, breaking local, ...",
    "favicon_url": "https://s.yimg.com/wm/mbr/images/yahoo-yep-favicon-v1.ico",
    "shortcut_icon_url": "https://s.yimg.com/wm/mbr/images/yahoo-yep-favicon-v1.ico",
    "viewport": "initial-scale=1, maximum-scale=1, user-scalable=0, shrink-to-fit..."
  },
  "requested_url": "https://login.yahoo.com/",
  "location": "https://login.yahoo.com/",
  "redirectLocations": null,
  "body": "<html id=\"Stencil\" lang=\"en-US\" class=\"js grid light-theme \"><he...",
  "initialBody": "<!DOCTYPE html>\n<html id=\"Stencil\" lang=\"en-US\" class=\"no-js gri...",
  "assets": [
    {
      "url": "https://ups.analytics.yahoo.com/ups/58824/sync?format=json&gdpr=...",
      "resourceType": "Fetch",
      "requestHeaders": {
        "Accept": "application/json",
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "access-control-allow-credentials": "true",
        "access-control-allow-origin": "https://login.yahoo.com",
        "age": "0",
        "content-type": "application/json",
        "date": "Wed, 26 Nov 2025 03:56:15 GMT",
        "p3p": "CP=NOI DSP COR LAW CURa DEVa TAIa PSAa PSDa OUR BUS UNI COM NAV",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-content-type-options": "nosniff"
      },
      "body": "{\"axid\": \"eS0zMVRoVFpWRTJ1R19nN1dId1NuR1VhQzZFaGNXNnR2Tn5B\"}"
    },
    {
      "url": "https://opus.analytics.yahoo.com/tag/opus-frame.html?referrer=ht...",
      "resourceType": "Document",
      "requestHeaders": {
        "Referer": "https://login.yahoo.com/",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "age": "48",
        "cache-control": "max-age=300",
        "content-encoding": "gzip",
        "content-security-policy": "default-src https:; script-src https: 'unsafe-inline'; style-src...",
        "content-type": "text/html",
        "date": "Wed, 26 Nov 2025 03:55:28 GMT",
        "etag": "W/\"f4613eae0fc38a92bed7c46fc8baeefc\"",
        "last-modified": "Tue, 14 Oct 2025 13:44:16 GMT",
        "server": "AmazonS3",
        "vary": "accept-encoding",
        "via": "1.1 c37f72766931ae9c3f146ffa54018d1c.cloudfront.net (CloudFront)",
        "x-amz-cf-id": "nE_3ar0XxI60YKSKlTKp3l6LfhYWU7k6bE2vff8hQ6TSCjjb9GDBwA==",
        "x-amz-cf-pop": "IAD89-C2",
        "x-amz-expiration": "expiry-date=\"Thu, 19 Nov 2026 00:00:00 GMT\", rule-id=\"standard-l...",
        "x-amz-server-side-encryption": "AES256",
        "x-cache": "Hit from cloudfront"
      },
      "body": "<!DOCTYPE html><html lang=\"en\"><head><meta http-equiv=\"Content-T..."
    },
    {
      "url": "https://ups.analytics.yahoo.com/ups/58746/sync?ui=12288b8a-c84e-...",
      "resourceType": "Image",
      "requestHeaders": {
        "Referer": "",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": null,
      "body": ""
    },
    {
      "url": "https://s.yimg.com/ss/analytics3.js",
      "resourceType": "Script",
      "requestHeaders": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "accept-ranges": "bytes",
        "age": "317",
        "ats-carp-promotion": "1, 1",
        "cache-control": "max-age=600",
        "content-encoding": "gzip",
        "content-length": "26272",
        "content-type": "application/javascript",
        "date": "Wed, 26 Nov 2025 03:50:58 GMT",
        "etag": "\"e9326f2295b4bcb1f5e00bd1beabd834-df\"",
        "last-modified": "Wed, 19 Nov 2025 15:52:02 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin, Accept-Encoding",
        "x-amz-id-2": "ljzdKpXqSyYpFhwNCBOMt0sxekat7QW/XWuHrRmqYgsrvcgMFg9BjjnShcpDmxrE...",
        "x-amz-request-id": "HKX10151GA9A7Y03",
        "x-amz-server-side-encryption": "AES256",
        "x-amz-version-id": "cbSPIsTp1M0ri3bgXk5_1tHUEmPorm9z",
        "x-content-type-options": "nosniff"
      },
      "body": "(function(De){typeof define==\"function\"&&define.amd?define(De):D..."
    },
    {
      "url": "https://ups.analytics.yahoo.com/ups/58699/cms?partner_id=SEMAS&o...",
      "resourceType": "Image",
      "requestHeaders": {
        "Referer": "",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": null,
      "body": ""
    },
    {
      "url": "https://consent.cmp.oath.com/cmp.js",
      "resourceType": "Script",
      "requestHeaders": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "age": "14",
        "cache-control": "max-age=3600",
        "content-encoding": "gzip",
        "content-type": "application/javascript",
        "date": "Wed, 26 Nov 2025 03:56:01 GMT",
        "etag": "W/\"4a32d9ba7441f8aae6da12a24a62bcfe\"",
        "last-modified": "Wed, 08 Oct 2025 17:34:03 GMT",
        "server": "AmazonS3",
        "vary": "accept-encoding",
        "via": "1.1 53b70ac9dc46d1c13992b291cf22a9aa.cloudfront.net (CloudFront)",
        "x-amz-cf-id": "rBL-Hp-Jm5d-zGpb80iQJF1ATmfW0wdff396HEFUJTIN2uQxRFCC-A==",
        "x-amz-cf-pop": "IAD12-P3",
        "x-amz-expiration": "expiry-date=\"Wed, 09 Oct 2030 00:00:00 GMT\", rule-id=\"webapp-sta...",
        "x-amz-server-side-encryption": "AES256",
        "x-cache": "Hit from cloudfront"
      },
      "body": "/*! CMP 7.0.2 Copyright 2018 Oath Holdings, Inc. */\n!function(){..."
    },
    {
      "url": "https://guce.yahoo.com/v1/consentRecord?consentTypes=iab%2CiabCC...",
      "resourceType": "XHR",
      "requestHeaders": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "Access-Control-Allow-Credentials": "true",
        "Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept, User-Agent, X-Fo...",
        "Access-Control-Allow-Methods": "HEAD, GET, OPTIONS",
        "Access-Control-Allow-Origin": "https://login.yahoo.com",
        "Connection": "keep-alive",
        "Content-Encoding": "gzip",
        "Content-Length": "131",
        "Content-Type": "application/json",
        "Date": "Wed, 26 Nov 2025 03:56:14 GMT",
        "Server": "guce",
        "Strict-Transport-Security": "max-age=31536000; includeSubDomains"
      },
      "body": "{\"identifier\":\"62tv745kicuiu\",\"identifierType\":\"bid\",\"tosRecords..."
    },
    {
      "url": "https://s.yimg.com/wm/mbr/2df03510d3a44bebaf3b767b14ccd6be88871c...",
      "resourceType": "Stylesheet",
      "requestHeaders": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "accept-ranges": "bytes",
        "age": "8144",
        "ats-carp-promotion": "1, 1",
        "cache-control": "public,max-age=31536000",
        "content-encoding": "gzip",
        "content-length": "140661",
        "content-type": "text/css",
        "date": "Wed, 26 Nov 2025 01:40:31 GMT",
        "etag": "\"c68492486108489b1f71307898c5317d-df\"",
        "last-modified": "Tue, 25 Nov 2025 22:29:15 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Accept-Encoding",
        "x-amz-id-2": "GZV1V9Bma5jc55qieSFFIzxo+QrTkuxhSF3ltD0tBogAKxSk4CAQRY5rGDSpKMLA...",
        "x-amz-request-id": "QE7Y5D42W80466XZ",
        "x-amz-server-side-encryption": "AES256",
        "x-content-type-options": "nosniff"
      },
      "body": "@font-face{font-family:\"Yahoo Sans\";font-display:block;src:url(h..."
    },
    {
      "url": "https://udc.yahoo.com/v2/public/yql?yhlVer=2&yhlClient=rapid&yhl...",
      "resourceType": "XHR",
      "requestHeaders": {
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "access-control-allow-credentials": "true",
        "access-control-allow-origin": "https://login.yahoo.com",
        "age": "0",
        "cache-control": "no-store, no-cache, private, max-age=0",
        "date": "Wed, 26 Nov 2025 03:56:14 GMT",
        "expires": "-1",
        "p3p": "policyref=\"http://info.yahoo.com/w3c/p3p.xml\", CP=\"CAO DSP COR C...",
        "pragma": "no-cache",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-envoy-upstream-service-time": "1"
      },
      "body": ""
    },
    {
      "url": "https://www.googletagmanager.com/gtag/js?id=G-P9C3W3ESF1",
      "resourceType": "Script",
      "requestHeaders": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KH..."
      },
      "responseHeaders": {
        "access-control-allow-credentials": "true",
        "access-control-allow-headers": "Cache-Control",
        "access-control-allow-origin": "*",
        "alt-svc": "h3=\":443\"; ma=2592000,h3-29=\":443\"; ma=2592000",
        "cache-control": "private, max-age=900",
        "content-encoding": "zstd",
        "content-length": "145405",
        "content-type": "application/javascript; charset=UTF-8",
        "cross-origin-resource-policy": "cross-origin",
        "date": "Wed, 26 Nov 2025 03:56:14 GMT",
        "expires": "Wed, 26 Nov 2025 03:56:14 GMT",
        "server": "Google Tag Manager",
        "strict-transport-security": "max-age=31536000; includeSubDomains",
        "vary": "Accept-Encoding",
        "x-xss-protection": "0"
      },
      "body": "\n// Copyright 2012 Google Inc. All rights reserved.\n \n(function(..."
    }
  ]
}
```
