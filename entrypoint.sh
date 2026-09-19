#!/bin/sh
set -eu

escape_javascript_string() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

sentry_frontend_dsn=$(escape_javascript_string "${SENTRY_FRONTEND_DSN:-}")
printf 'window.appConfig = {"SENTRY_FRONTEND_DSN":"%s"};\n' "$sentry_frontend_dsn" > public/config.js

exec /go/bin/server
