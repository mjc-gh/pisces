# pisces

[![.github/workflows/ci.yml](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml/badge.svg)](https://github.com/mjc-gh/pisces/actions/workflows/ci.yml)
[![GitHub
Release](https://img.shields.io/github/v/release/mjc-gh/pisces?label=Latest%20Release)](https://github.com/mjc-gh/pisces/releases)

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
   --concurrency int, -c int   (default: 0)
   --port int                  (default: 9222)
   --device-type string        (default: "desktop")
   --device-size string        (default: "large")
   --host string               (default: "127.0.0.1")
   --user-agent string         (default: "chrome")
```

See command specific help for output flags.

### Result Output

Below is an example of the `analyze` task's output for a single visit to `login.yahoo.com`. Some of the assets and response details have been truncated for brevity.

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
          "value": "{\"language\":\"en-US\",\"colorDepth\":24,\"deviceMemory\":8,\"pixelRatio\":1,\"hardwareConcurrency\":10,\"timezoneOffset\":300,\"timezone\":\"America/New_York\",\"sessionStorage\":1,\"localStorage\":1,\"indexedDb\":1,\"cpuClass\":\"unknown\",\"platform\":\"MacIntel\",\"doNotTrack\":\"unknown\",\"plugins\":{\"count\":5,\"hash\":\"2c14024bf8584c3f7f63f24ea490e812\"},\"canvas\":\"canvas winding:yes~canvas\",\"webgl\":1,\"webglVendorAndRenderer\":\"Google Inc. (Apple)~ANGLE (Apple, ANGLE Metal Renderer: Apple M4, Unspecified Version)\",\"adBlock\":0,\"hasLiedLanguages\":0,\"hasLiedResolution\":0,\"hasLiedOs\":1,\"hasLiedBrowser\":0,\"touchSupport\":{\"points\":0,\"event\":0,\"start\":0},\"fonts\":{\"count\":27,\"hash\":\"d52a1516cfb5f1c2d8a427c14bc3645f\"},\"audio\":\"124.04347745512496\",\"resolution\":{\"w\":\"800\",\"h\":\"600\"},\"availableResolution\":{\"w\":\"600\",\"h\":\"800\"},\"ts\":{\"serve\":1764736582612,\"render\":1764736582902}}"
        },
        {
          "name": "crumb",
          "type": "hidden",
          "value": "UF/63CgZqrqctVz4oFZYcg"
        },
        {
          "name": "acrumb",
          "type": "hidden",
          "value": "X4mBJNPf"
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
      "href": "https://help.yahoo.com/kb/index?locale=en_US&page=product&y=PROD_ACCT",
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
      "href": "https://login.yahoo.com/account/create?specId=yidregsimplified&done=https%3A%2F%2Fwww.yahoo.com",
      "text": "Create an account",
      "class": "pure-button puree-button-secondary challenge-button"
    }
  ],
  "head": {
    "title": "Login - Sign in to Yahoo",
    "description": "Sign in to access the best in class Yahoo Mail, breaking local, national and global news, finance, sports, music, movies... You get more out of the web, you get more out of life.",
    "favicon_url": "https://s.yimg.com/wm/mbr/images/yahoo-yep-favicon-v1.ico",
    "shortcut_icon_url": "https://s.yimg.com/wm/mbr/images/yahoo-yep-favicon-v1.ico",
    "viewport": "initial-scale=1, maximum-scale=1, user-scalable=0, shrink-to-fit=no"
  },
  "visible_text": "Help Terms Privacy Sign in using your Yahoo account Username, em...",
  "requested_url": "https://login.yahoo.com/",
  "location": "https://login.yahoo.com/",
  "redirect_locations": null,
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
  "body": "<html id=\"Stencil\" lang=\"en-US\" c",
  "initial_body": "<!DOCTYPE html>\n<html id=\"Stencil",
  "assets": [
    {
      "url": "https://opus.analytics.yahoo.com/",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "opus.analytics.yahoo.com",
        "valid_from": "2025-09-23T20:00:00-04:00",
        "valid_to": "2025-12-24T18:59:59-05:00",
        "sans": [
          "opus.analytics.yahoo.com",
          "prod.opus.aolp-ds-prd.aws.oath.cloud"
        ]
      },
      "resource_type": "Document",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
      },
      "response_headers": {
        "age": "271",
        "cache-control": "max-age=300",
        "content-encoding": "gzip",
        "content-security-policy": "default-src https:; script-src https: 'unsafe-inline'; style-src https: 'unsafe-inline'",
        "content-type": "text/html",
        "date": "Wed, 03 Dec 2025 04:31:53 GMT",
        "etag": "W/\"f4613eae0fc38a92bed7c46fc8baeefc\"",
        "last-modified": "Tue, 14 Oct 2025 13:44:16 GMT",
        "server": "AmazonS3",
        "vary": "accept-encoding",
        "via": "1.1 79bf0c7fd285682183c6a826f8dfe31e.cloudfront.net (CloudFront)",
        "x-amz-cf-id": "ZYRlvwBsjbtTfSTHT2QYx5b5ow3uozMYClodK9GuQP4z04LWKQpMMA==",
        "x-amz-cf-pop": "JFK50-P14",
        "x-amz-expiration": "expiry-date=\"Thu, 19 Nov 2026 00:00:00 GMT\", rule-id=\"standard-lifecycle\"",
        "x-amz-server-side-encryption": "AES256",
        "x-cache": "Hit from cloudfront"
      },
      "response_status": 200,
      "body": "<!DOCTYPE html><html lang=\"en\"><h"
    },
    {
      "url": "https://ups.analytics.yahoo.com/u",
      "resource_type": "Image",
      "request_headers": {
        "Referer": "",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
      },
      "response_headers": null,
      "initiator_url": "https://login.yahoo.com/"
    },
    {
      "url": "https://opus.analytics.yahoo.com/",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "opus.analytics.yahoo.com",
        "valid_from": "2025-09-23T20:00:00-04:00",
        "valid_to": "2025-12-24T18:59:59-05:00",
        "sans": [
          "opus.analytics.yahoo.com",
          "prod.opus.aolp-ds-prd.aws.oath.cloud"
        ]
      },
      "resource_type": "Script",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
      },
      "response_headers": {
        "age": "169",
        "cache-control": "max-age=300",
        "content-encoding": "gzip",
        "content-security-policy": "default-src https:; script-src https: 'unsafe-inline'; style-src https: 'unsafe-inline'",
        "content-type": "application/javascript",
        "date": "Wed, 03 Dec 2025 04:33:34 GMT",
        "etag": "W/\"5d5fd12b6bd2f3056713173b99d51c4f\"",
        "last-modified": "Tue, 14 Oct 2025 13:44:16 GMT",
        "server": "AmazonS3",
        "vary": "accept-encoding",
        "via": "1.1 79bf0c7fd285682183c6a826f8dfe31e.cloudfront.net (CloudFront)",
        "x-amz-cf-id": "qKeFwrH3J4gNk99Cv2HiOts8rdwdnp5zf5xliL011RyQkos5BqkA4g==",
        "x-amz-cf-pop": "JFK50-P14",
        "x-amz-expiration": "expiry-date=\"Thu, 19 Nov 2026 00:00:00 GMT\", rule-id=\"standard-lifecycle\"",
        "x-amz-server-side-encryption": "AES256",
        "x-cache": "Hit from cloudfront"
      },
      "response_status": 200,
      "initiator_url": "https://login.yahoo.com/",
      "body": "(()=>{\"use strict\";var n=function"
    },
    {
      "url": "https://ups.analytics.yahoo.com/u",
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
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
      },
      "response_headers": {
        "access-control-allow-credentials": "true",
        "access-control-allow-origin": "https://login.yahoo.com",
        "age": "0",
        "content-type": "application/json",
        "date": "Wed, 03 Dec 2025 04:36:23 GMT",
        "p3p": "CP=NOI DSP COR LAW CURa DEVa TAIa PSAa PSDa OUR BUS UNI COM NAV",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "body": "{\"axid\": \"y-wHs6xtdE2uLXOHSbaN5Lu"
    },
    {
      "url": "https://s.yimg.com/wm/mbr/2df0351",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert Global G2 TLS RSA SHA256 2020 CA1",
        "subject_name": "*.www.yahoo.com",
        "valid_from": "2025-11-13T19:00:00-05:00",
        "valid_to": "2025-12-31T18:59:59-05:00",
        "sans": [
          "*.www.yahoo.com",
          "*.yahoo.com",
          "ymail.com",
          "s.yimg.com"
        ]
      },
      "resource_type": "Stylesheet",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
      },
      "response_headers": {
        "accept-ranges": "bytes",
        "age": "615352",
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
        "x-amz-id-2": "GZV1V9Bma5jc55qieSFFIzxo+QrTkuxhSF3ltD0tBogAKxSk4CAQRY5rGDSpKMLAT+S0UK7O/yc=",
        "x-amz-request-id": "QE7Y5D42W80466XZ",
        "x-amz-server-side-encryption": "AES256",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "initiator_url": "https://login.yahoo.com/",
      "body": "@font-face{font-family:\"Yahoo San"
    },
    {
      "url": "https://s.yimg.com/bw/fonts/yahoo",
      "certificate_info": {
        "protocol": "TLS 1.3",
        "issuers": "DigiCert Global G2 TLS RSA SHA256 2020 CA1",
        "subject_name": "*.www.yahoo.com",
        "valid_from": "2025-11-13T19:00:00-05:00",
        "valid_to": "2025-12-31T18:59:59-05:00",
        "sans": [
          "*.www.yahoo.com",
          "*.yahoo.com",
          "ymail.com",
          "s.yimg.com"
        ]
      },
      "resource_type": "Font",
      "request_headers": {
        "Origin": "https://login.yahoo.com",
        "Referer": "https://s.yimg.com/wm/mbr/2df03510d3a44bebaf3b767b14ccd6be88871cfc/yahoo-main.css",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
      },
      "response_headers": {
        "accept-ranges": "bytes",
        "access-control-allow-origin": "*",
        "age": "280742",
        "ats-carp-promotion": "1, 1",
        "cache-control": "public, max-age=604800",
        "content-length": "34588",
        "content-type": "font/woff2",
        "date": "Sat, 29 Nov 2025 22:37:21 GMT",
        "etag": "\"492a0a160b8da9414134282ef8b62f78\"",
        "last-modified": "Fri, 14 Feb 2025 01:59:26 GMT",
        "referrer-policy": "no-referrer-when-downgrade",
        "server": "ATS",
        "strict-transport-security": "max-age=31536000",
        "vary": "Origin",
        "x-amz-id-2": "ieUBmgK49fqSaIPnsHM9tYFIu0CdQKCDM+0UhxXQNzeDcOmAm8SDvIaoWj2gPYxgmkUUyCO5U2I=",
        "x-amz-request-id": "GBNAGTQVZZGVP4CF",
        "x-amz-server-side-encryption": "AES256",
        "x-amz-version-id": "wUt5QWMHFonqO_Z13wqeP1wud1inlbiC",
        "x-content-type-options": "nosniff"
      },
      "response_status": 200,
      "initiator_url": "https://s.yimg.com/wm/mbr/2df0351",
      "body": "wOF2\u0000\u0001\u0000\u0000\u0000\u0000�\u001c\u0000\u0011\u0000\u0000\u0000\u0001\u0015\u0014\u0000\u0000��\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000"
    },
    {
      "url": "https://guce.yahoo.com/v1/consent",
      "certificate_info": {
        "protocol": "TLS 1.2",
        "issuers": "DigiCert SHA2 High Assurance Server CA",
        "subject_name": "guce.oath.com",
        "valid_from": "2025-08-20T20:00:00-04:00",
        "valid_to": "2026-02-11T18:59:59-05:00",
        "sans": [
          "guce.oath.com",
          "guce.yahoo.com",
          "guce.aol.com",
          "guce.engadget.com"
        ]
      },
      "resource_type": "XHR",
      "request_headers": {
        "Referer": "https://login.yahoo.com/",
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
      },
      "response_headers": {
        "Access-Control-Allow-Credentials": "true",
        "Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept, User-Agent, X-Forwarded-For, X-Oath-Gcrumb",
        "Access-Control-Allow-Methods": "HEAD, GET, OPTIONS",
        "Access-Control-Allow-Origin": "https://login.yahoo.com",
        "Connection": "keep-alive",
        "Content-Encoding": "gzip",
        "Content-Length": "131",
        "Content-Type": "application/json",
        "Date": "Wed, 03 Dec 2025 04:36:23 GMT",
        "Server": "guce",
        "Strict-Transport-Security": "max-age=31536000; includeSubDomains"
      },
      "response_status": 200,
      "body": "{\"identifier\":\"fn1jb6pkivfi6\",\"id"
    }
  ]
}
```
