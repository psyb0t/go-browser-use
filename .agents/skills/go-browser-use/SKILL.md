---
name: go-browser-use
description: "Write Go code using the unofficial Browser Use Cloud API v4 client when asked to manage Browser Use agent runs, sessions, workspaces, browsers, profiles, files, or integrations, or when a project imports github.com/psyb0t/go-browser-use."
homepage: https://github.com/psyb0t/go-browser-use
user-invocable: true
metadata:
  openclaw:
    emoji: "🌐"
    requires:
      bins:
        - go
      env:
        - BROWSER_USE_API_KEY
permissions:
  filesystem: "Guides edits to Go source in the current project. The library does not read or write local files."
  shell: "The Go toolchain to add and compile the dependency through the project's build commands."
  network: "Module downloads use the Go module proxy. API requests use api.browser-use.com; file transfers use provider-issued presigned URLs without the API key."
---

# go-browser-use

Unofficial, typed Go client for the Browser Use Cloud API v4, in package `browseruse` at the module root. Read [references/setup.md](references/setup.md) before writing an integration, for authentication, operation names, pagination, and response handling.

## Security & safety

- Load `BROWSER_USE_API_KEY` from the caller's environment or secret store. Set `X-Browser-Use-API-Key` in a `WithRequestEditorFn`. Never hardcode or log credentials, task contents, secret bindings, CDP URLs, or presigned URLs.
- The API base URL is `https://api.browser-use.com/api/v4`. Use an HTTP client with a timeout or a context deadline.
- Starting runs or browsers can incur charges. Cancellation, deletion, purging, sharing, uploads, and integration changes affect account state. Perform those actions only when the user requested them.
- Presigned file transfers use a separate HTTP request. Do not forward the Browser Use API key to the storage host.

## When to use

- A Go project needs Browser Use Cloud v4 agent runs, session queues, files, browser sessions, profiles, or integrations.
- The project already imports `github.com/psyb0t/go-browser-use`.

## When not to use

- The caller needs a non-Go or provider-maintained SDK.
- The task targets API v2 or v3; this module uses v4 only.
- The task changes the generated client. Read the repository's `AGENTS.md` for the generation pipeline instead.

## Workflow

1. Add the module through the project's dependency workflow and read its current generated signatures.
2. Construct `browseruse.NewClientWithResponses` with the v4 base URL, `WithRequestEditorFn` authentication, and appropriate HTTP timeouts.
3. Use the exact generated operation and request types listed in the reference. Check the returned error, HTTP status, and decoded status field before reading it.
4. Follow pagination explicitly. Poll run status until terminal, then fetch the run result. Do not start duplicate paid runs as a retry strategy.
5. Verify the integration against an `httptest` server. Stop when the requested call compiles and its method, path, auth, request body, and response handling pass the caller's tests.

The README usage example also lives in `example_test.go`. Optional nil pointers omit fields; explicit JSON null resets require a raw-body method. Cost responses stay decimal strings.
