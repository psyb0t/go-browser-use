// Package browseruse is an unofficial Go client for the Browser Use Cloud API.
package browseruse

//go:generate bash scripts/fetch-spec.sh
//go:generate bash scripts/patch-spec.sh
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.fixed.json
