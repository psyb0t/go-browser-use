# Changelog

Changes are listed per release. Versions follow [semver](https://semver.org).

## v0.1.3, 2026-10-27

- The development image is based on Debian (`golang:1.26.10-trixie`) instead of Alpine. Alpine keeps only the latest build of each package, so the exact pins broke the image build whenever a package was updated. System packages now come from the snapshot.debian.org date the base image was built from, which keeps them fixed for a given base image digest. jq 1.8.2, shellcheck 0.11.0 and shfmt 3.13.1 keep their versions, and Semgrep runs in its own virtual environment.
- Raised the minimum Go version in `go.mod` from 1.26.9 to 1.26.10, which fixes the `crypto/x509` vulnerability GO-2026-6809 found by govulncheck.
- The issue-pull workflow runs at hours derived from the repository name instead of the same four hours as every other repository.

## v0.1.2, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.1, 2026-10-17

- Generated doc comments keep the API's own description text. The spec patch no longer rewrites dashes or wording in descriptions.
- `make sec` uses Semgrep 1.181.0. Its pinned requirements now hash every published file, so the development image also builds on arm64.

## v0.1.0, 2026-10-17

- Added `browseruse`, a REST client generated with oapi-codegen from Browser Use Public API v4, spec version 4.0.0, OpenAPI 3.1.0. It covers all 50 operations across runs, sessions, workspaces, files, wallets, integrations, extensions, browsers, and profiles.
- `make generate` downloads `https://docs.browser-use.com/cloud/openapi/v4.json`, converts nullable unions, constants, and exclusive numeric bounds to OpenAPI 3.0.3 forms, normalizes description prose, and regenerates the client. The spec is not committed.
- Unformatted number fields use `float64`; decimal cost strings stay strings. Added HTTP contract tests, the usage example, and agent integrations.
