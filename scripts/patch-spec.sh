#!/bin/bash
# Downconverts JSON Schema constructs to OpenAPI 3.0 for oapi-codegen:
# null union branches become nullable, const becomes a singleton enum, and
# numeric exclusive bounds become a bound plus the boolean exclusive flag.
# Union alternatives other than null remain intact. No operations are removed.
# Description punctuation and provider first-person prose are normalized for
# generated Go comments only; schema names and wire values stay unchanged.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

on_error() {
	local exit_code="$1"
	local line="$2"
	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

tmp="$(mktemp "${SPEC_DIR}/.openapi.fixed.XXXXXX")"
trap 'rm -f -- "$tmp"' EXIT

jq '
  walk(if type == "object" then
    (if (.description | type) == "string" then
      .description |= (gsub(" *[—–] *"; ", ")
        | gsub("We recommend that you create"; "Create"))
    else . end)
    |
    (if has("anyOf") and any(.anyOf[]; .type == "null") then
      .anyOf |= map(select(.type != "null"))
      | .nullable = true
      | if (.anyOf | length) == 1 then
          . + .anyOf[0] | del(.anyOf)
        else . end
    else . end)
    | (if has("const") then .enum = [.const] | del(.const) else . end)
    | (if (.exclusiveMinimum | type) == "number" then
        .minimum = .exclusiveMinimum | .exclusiveMinimum = true
      else . end)
    | (if (.exclusiveMaximum | type) == "number" then
        .maximum = .exclusiveMaximum | .exclusiveMaximum = true
      else . end)
  else . end)
  | .openapi = "3.0.3"
' "${SPEC_DIR}/openapi.json" >"$tmp"
chmod 0644 "$tmp"
mv -- "$tmp" "${SPEC_DIR}/openapi.fixed.json"
log INFO "patched OpenAPI document for generation"
