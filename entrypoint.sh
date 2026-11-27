#!/bin/sh
set -eu

escape_javascript_string() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

sentry_frontend_dsn=$(escape_javascript_string "${SENTRY_FRONTEND_DSN:-}")
service_version=$(escape_javascript_string "${SERVICE_VERSION:-dev}")
printf 'window.appConfig = {"SENTRY_FRONTEND_DSN":"%s","SERVICE_VERSION":"%s"};\n' "$sentry_frontend_dsn" "$service_version" > public/config.js

exec /go/bin/server
