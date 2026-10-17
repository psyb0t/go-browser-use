# Changelog

Changes are listed per release. Versions follow [semver](https://semver.org).

## v0.1.0, 2026-10-17

- Added `browseruse`, a REST client generated with oapi-codegen from Browser Use Public API v4, spec version 4.0.0, OpenAPI 3.1.0. It covers all 50 operations across runs, sessions, workspaces, files, wallets, integrations, extensions, browsers, and profiles.
- `make generate` downloads `https://docs.browser-use.com/cloud/openapi/v4.json`, converts nullable unions, constants, and exclusive numeric bounds to OpenAPI 3.0.3 forms, normalizes description prose, and regenerates the client. The spec is not committed.
- Unformatted number fields use `float64`; decimal cost strings stay strings. Added HTTP contract tests, the usage example, and agent integrations.
