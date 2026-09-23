# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Optional ECS-shaped JSON logging for API requests and URL query outcomes
- Validated request correlation IDs returned through `X-Request-ID`
- Service version and environment metadata for structured logs
- ECS version and event outcome fields to all structured logs
- Timestamp index for query history and metrics lookups
- Generic OIDC authentication with discovery, PKCE, validated ID tokens, and provider logout
- Shared authenticated sessions, frontend user identity, and logout for all authentication providers

### Changed

- Authentication providers are selected explicitly with `AUTH_PROVIDER` and conflicting configuration is rejected at startup
- Azure AD and Okta now use provider-neutral session identity and per-login state

## [1.0.0] - 2026-09-19

### Added

- Query history tracking with timestamps and success/failure indicators
- Most Wanted tab displaying top 10 unresolved URL keys by request count
- Edit functionality for Most Wanted entries to pre-populate the add-URL modal
- Metrics panel showing all-time and seven-day query statistics
- Seven-day stacked histogram of successful and failed queries
- Dark mode toggle with improved color palette and card styling
- Delete functionality for URL entries with confirmation modal
- Support for spaces in URL keys
- Okta authentication integration
- Slack command parameter support
- Sentry error tracking for API and frontend
- IP CIDR whitelist support for authentication
- X-Forwarded-For trust level configuration
- Unit tests in CircleCI workflow
- Auth redirect to original URL after login
- CONTRIBUTING.md guide
- CircleCI status badge in README

### Changed

- Modernized Docker build to use Go 1.24.1, Node.js 22.14.0, and Alpine 3.21
- Improved Docker dependency caching with BuildKit cache mounts
- Docker container now runs as non-root `app` user
- Replaced Bash entrypoint with POSIX-compatible sh script
- Made frontend runtime configuration injection idempotent
- Added graceful shutdown handling for SIGTERM and SIGINT
- Updated CircleCI to run only on tag pushes
- CircleCI now publishes versioned Docker images (e.g., `v1.2.3`) and `latest`
- Upgraded CircleCI executor to `cimg/go:1.24.1-node`
- Enabled Docker BuildKit in CircleCI for cache-mount support
- Reordered main tabs to: Most Popular, Most Wanted, History
- Migrated from gopkg.toml to Go modules (go.mod)
- Updated Material-UI icons and imports
- Improved header text contrast for dark mode
- Removed Cypress test suite references
- Refreshed README documentation

### Fixed

- Slack bot command message when key not found
- Timezone handling in metrics display
- Proxy middleware and public path configuration
- Typescript errors and type configurations
- Prettier formatting issues
- Yarn lock file inconsistencies
- Dockerfile network timeout issues
- Auth redirect URL construction
- Cookie same-site mode to use lax for link compatibility
- Auth cookie security with custom expiry
- Opensearch.xml XML header

### Security

- Bumped axios from 0.21.0 to 0.21.1
- Bumped node-notifier from 8.0.0 to 8.0.1
- Bumped immer to 8.0.1
- Bumped tar from 6.1.0 to 6.1.11
- Bumped path-parse from 1.0.6 to 1.0.7
- Bump tmpl from 1.0.4 to 1.0.5
- Bumped url-parse from 1.5.1 to 1.5.3
- Bumped color-string from 1.5.4 to 1.5.5
- Bumped dns-packet from 1.3.1 to 1.3.4
- Bumped acorn from 5.7.3 to 5.7.4
- Bumped lodash.template from 4.4.0 to 4.5.0
- Bumped eslint-utils from 1.3.1 to 1.4.2
- Bumped mixin-deep from 1.3.1 to 1.3.2
- Fixed security issue with set-value
- Multiple golang and node Alpine version bumps

### Removed

- Cypress e2e test suite references
- Legacy gopkg.toml dependency management
- Redundant Docker runtime packages (bash)
