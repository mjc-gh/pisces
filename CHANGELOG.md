# Changelog

## Unreleased

Changes that make the `engine` package safe to embed in a long running
service. New APIs are additive and the existing CLI, `pisces-web` and the
exported function signatures keep working. The behavior changes are listed
under "Changed".

### Added

- `engine.WithExecAllocatorOptions(opts ...chromedp.ExecAllocatorOption)`
  passes extra options (`chromedp.NoSandbox`, `chromedp.ExecPath`,
  `chromedp.Flag(...)`, ...) to the local exec allocator. They are appended
  after pisces' defaults, so they win. They are ignored, with a warning, when
  `WithRemoteAllocator` is used.
- `engine.WithTaskTimeout(d)` limits how long a task runs once a worker starts
  it. On expiry the task's browser is released and `Result.Error` wraps
  `context.DeadlineExceeded`. Zero means no timeout (the previous behavior).
- `Asset.RawBody` and `Visit.RawInitialBody` hold the exact response bytes
  (not part of the JSON output).
- `content_type` and `size` on assets, and `initial_content_type` and
  `initial_size` on the visit, in the JSON output.
- `full_page` parameter for the `screenshot` action to capture the whole
  scrollable page. The default is unchanged.

### Changed

- **`Result.Error` is now serialized as its message.** The `error` interface
  was previously encoded as `{}`, so consumers of the JSON saw `"error":{}` for
  failed results and now see `"error":"<message>"`. A successful result still
  has no `error` key.
- **Binary bodies are no longer included in the JSON.** `body` (assets) and
  `initial_body` are only set when the content type is textual (`text/*`,
  JSON, XML, JavaScript, SVG) and the bytes are valid UTF-8. Previously every
  body was converted with `string(bytes)`, which corrupted binary data when
  encoded as JSON. Text-only pages keep their `body` values. The JSON shape for
  a text-only page is otherwise unchanged apart from the new `content_type`,
  `size`, `initial_content_type` and `initial_size` keys. A text body that is not
  valid UTF-8 (for example ISO-8859-1) is now omitted from the JSON but is
  available in `RawBody`.
- `initial_body_length` counts the bytes of the raw initial body, which is
  identical to before for text documents.
- A visit now waits for every in-flight response body fetch to finish (or for
  the context to be done) before it is assembled. Previously the fetches wrote
  to the visit from other goroutines after it had been returned, which was a
  data race and could drop bodies.
- `assets` are listed in the order the requests were first seen. They were
  previously in random map order.
- An `analyze` task whose context is cancelled or times out now reports it as
  an error (wrapping the context error) instead of success, and returns the
  partial payload it collected in `Result.Result` alongside the error.
- Cancelling the context passed to `Engine.Start` ends in-flight tasks,
  completes queued tasks with an error without launching a browser, and no
  longer leaves `Engine.Shutdown` blocked sending to a results channel that no
  one is reading (results that nobody receives are dropped once the context is
  cancelled).

### Notes for embedders

- Under the local exec allocator every task starts its own Chrome process,
  which exits when the task finishes. Concurrency therefore bounds the number
  of browsers.
- Do not pass `chromedp.UserDataDir` with a concurrency above 1: the browsers
  would share one profile and all but one fail to start. Leave it unset and set
  `TMPDIR` instead; chromedp creates and removes a private profile directory for
  each task.
