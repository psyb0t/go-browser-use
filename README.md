# go-browser-use

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-browser-use.svg)](https://pkg.go.dev/github.com/psyb0t/go-browser-use)
[![CI](https://github.com/psyb0t/go-browser-use/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-browser-use/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-browser-use/badges/version.svg)](https://github.com/psyb0t/go-browser-use/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-browser-use/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-browser-use/badges/importers.svg)](https://github.com/psyb0t/go-browser-use/blob/badges/importers.md)

Unofficial Go client for the [Browser Use Cloud API](https://docs.browser-use.com), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the public v4 OpenAPI document. I am not affiliated with Browser Use.

The package is `browseruse`, in the repository root. Everything in `client.gen.go` is generated. Don't edit it by hand, regenerate it.

## Install

```sh
go get github.com/psyb0t/go-browser-use
```

## Usage

Every endpoint in the spec has a typed method on `ClientWithResponses`. Responses decode into `JSON200` and the other status fields.

```go
package main

import (
	"context"
	"net/http"
	"os"

	browseruse "github.com/psyb0t/go-browser-use"
)

func main() {
	auth := browseruse.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"X-Browser-Use-API-Key",
				os.Getenv("BROWSER_USE_API_KEY"),
			)

			return nil
		},
	)

	client, err := browseruse.NewClientWithResponses(
		"https://api.browser-use.com/api/v4",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.ListRunsRunsGetWithResponse(
		context.Background(),
		&browseruse.ListRunsRunsGetParams{},
	)
	if err != nil {
		panic(err)
	}

	if res.JSON200 == nil {
		panic(res.Status())
	}

	_ = res.JSON200.Runs
}
```

The same code is in `example_test.go`, so `make test` checks that it compiles.

Method names retain the spec's operation IDs, including resource and HTTP method suffixes. `CreateRunRunsPostWithResponse` starts a run. The client covers runs, events, sessions, queues, sharing, workspaces, files, wallets, integrations, extensions, browsers, and profiles.

The full API is in the [Go reference](https://pkg.go.dev/github.com/psyb0t/go-browser-use); after `go get`, `go doc github.com/psyb0t/go-browser-use` prints it locally.

## Auth

Browser Use uses the `APIKeyHeader` security scheme: set `X-Browser-Use-API-Key` through `WithRequestEditorFn`, as in the example. Read the key from `BROWSER_USE_API_KEY`; never hardcode or log it. The production base URL is `https://api.browser-use.com/api/v4`, including the version prefix.

The error return covers transport and decoding failures. For API errors, check `StatusCode()`, the decoded status field when present, and `Body`. Configure request deadlines or pass a timed `http.Client` through `WithHTTPClient`. No retries or pagination loops run automatically.

## Regenerating

The spec is not committed. `make generate` downloads the current OpenAPI document into the gitignored `api/` directory and regenerates the client:

```sh
make generate
```

`gen.go` declares the steps, so `make generate` is just `go generate ./...`:

- `scripts/fetch-spec.sh` downloads `https://docs.browser-use.com/cloud/openapi/v4.json` to `api/openapi.json`.
- `scripts/patch-spec.sh` writes `api/openapi.fixed.json` in OpenAPI 3.0.3 form. It converts null union branches to `nullable`, `const` to singleton enums, and numeric exclusive bounds to a bound plus an exclusive flag. Other union alternatives and all operations remain intact.
- oapi-codegen runs with `config.yaml`, mapping unformatted `number` fields to `float64`. Decimal cost strings remain strings.

Optional nil pointer fields are omitted. To send an explicit null, such as clearing a session title, use the operation's `WithBodyWithResponse` method with a JSON reader. Upload endpoints return presigned URLs; send file bytes separately without the Browser Use API key. Event retrieval returns JSON pages, not a WebSocket stream.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-browser-use) teaches a coding agent how to use go-browser-use when it writes Go. It works in any agent that reads `.agents/skills/`, and installs natively in the clients below.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-browser-use@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-browser-use@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

The release workflow publishes the skill to ClawHub:

```bash
openclaw skills install @psyb0t/go-browser-use
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs only Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the spec and regenerate
```

## License

[MIT](LICENSE). The generated code comes from Browser Use's public v4 API specification. The specification itself is not redistributed here and has no license or terms URL in its metadata.

Vendored development tools retain their own licenses, including GPL and MPL notices.
