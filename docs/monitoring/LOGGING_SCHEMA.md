# API Logging Schema

This document describes the JSON schema for application logs written to stdout. `JSON_LOGS=true` emits structured JSON for request, URL query, frontend, lifecycle, database, authentication, Slack, and background-task events. Audit events are always emitted as structured JSON, regardless of `JSON_LOGS`.

The application emits Elastic Common Schema (ECS)-shaped JSON because Elastic Agent is the expected initial consumer. The selected fields either match OpenTelemetry semantic conventions directly or have a documented mapping to the OpenTelemetry Logs Data Model. Fields specific to `go-url` use the `go_url.*` namespace.

ECS and OpenTelemetry semantic conventions are converging, but they are not identical. The application does not emit duplicate ECS and OpenTelemetry representations of the same value. Translation to OTLP can be performed by an OpenTelemetry Collector or Elastic ingest pipeline.

See the [Elastic Agent example](ELASTIC_AGENT.md) and [OpenTelemetry Collector example](OPENTELEMETRY_COLLECTOR.md) for collection and field-mapping configurations.

## ECS and OpenTelemetry mapping

| JSON field | ECS status | OpenTelemetry destination | Description |
| --- | --- | --- | --- |
| `@timestamp` | native | `Timestamp` | Time the event occurred, in RFC 3339 format with nanosecond precision |
| `ecs.version` | native | record attribute | ECS version followed by the event; currently `9.5.0` |
| `log.level` | native | `SeverityText` | Text severity: `DEBUG`, `INFO`, `WARN`, or `ERROR` |
| `message` | native | `Body` | Short, stable human-readable description |
| `event.action` | native | `EventName` | Stable machine-readable event type |
| `event.kind` | native | record attribute | `event` for dedicated audit records |
| `event.dataset` | native | record attribute | `go_url.audit` identifies dedicated audit records |
| `event.category` | native | record attribute | High-level ECS event category array |
| `event.type` | native | record attribute | ECS event type array |
| `event.outcome` | native | record attribute | Whether the operation succeeded or failed |
| `event.duration` | native | record attribute | Event duration in nanoseconds |
| `service.name` | exact match | resource attribute | Service emitting the record |
| `service.version` | exact match | resource attribute | Released application version |
| `service.environment` | native | `deployment.environment.name` resource attribute | Deployment environment |
| `http.request.id` | native | record attribute | Request correlation identifier |
| `application.component` | custom | record attribute | Startup component associated with a lifecycle event |
| `server.port` | custom | record attribute | HTTP port used by the application |
| `database.system` | custom | record attribute | Database system, currently `postgresql` |
| `database.operation` | custom | record attribute | Stable database initialization or cleanup step |
| `auth.provider` | custom | record attribute | Configured authentication provider |
| `slack.operation` | custom | record attribute | Slack command, connection, or message operation |
| `background.task` | custom | record attribute | Stable name of a background task |
| `frontend.component` | custom | record attribute | Allowlisted frontend component associated with a reported event |
| `source.ip` | native | record attribute | Client IP attributed to a query, frontend, or audit event |
| `user.name` | native | record attribute | Authenticated username from the server-side session |
| `user.id` | native | record attribute | Stable authenticated subject from the identity provider |
| `go_url.url.key` | custom | record attribute | URL key affected by an audited management action |
| `error.type` | native | record attribute | Stable error category; never a raw error string |
| `trace.id` | native | `TraceId` | W3C trace identifier when tracing is introduced |
| `span.id` | native | `SpanId` | W3C span identifier when tracing is introduced |
| All other fields | native, matching, or custom | record attributes | Event-specific attributes |

The ingestion pipeline should derive OpenTelemetry `SeverityNumber` from `log.level`. Observation time should be assigned by the collector rather than emitted by the application.

## Common fields

Every record contains:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `@timestamp` | string | yes | UTC event time in RFC 3339 Nano format |
| `ecs.version` | string | yes | ECS version followed by the event; `9.5.0` |
| `log.level` | string | yes | `DEBUG`, `INFO`, `WARN`, or `ERROR` |
| `message` | string | yes | Stable human-readable event description |
| `event.action` | string | yes | Stable event identifier |
| `event.outcome` | string | yes | `success` or `failure` |
| `event.kind` | string | no | `event` on dedicated audit records |
| `event.dataset` | string | no | `go_url.audit` on dedicated audit records |
| `service.name` | string | yes | `go-url-api` |
| `service.version` | string | no | Released application version |
| `service.environment` | string | no | Deployment environment such as `production` |
| `http.request.id` | string | no | Validated or server-generated request correlation ID |
| `application.component` | string | no | Startup component associated with a lifecycle event |
| `server.port` | integer | no | HTTP port used by the application |
| `database.system` | string | no | Database system, currently `postgresql` |
| `database.operation` | string | no | Stable database initialization or cleanup step |
| `auth.provider` | string | no | Configured authentication provider |
| `slack.operation` | string | no | Slack command, connection, or message operation |
| `background.task` | string | no | Stable name of a background task |
| `frontend.component` | string | no | Allowlisted frontend component associated with a reported event |
| `source.ip` | string | no | Client IP attributed to query, frontend, or audit events; query and audit events use the direct connection unless trusted forwarded-header handling is enabled |
| `user.name` | string | no | Authenticated username from the server-side session |
| `user.id` | string | no | Stable authenticated subject from the identity provider |
| `go_url.url.key` | string | no | URL key affected by an audited management action |
| `error.type` | string | no | Stable error category, not raw error text |
| `trace.id` | string | no | W3C trace ID when distributed tracing is introduced |
| `span.id` | string | no | W3C span ID when distributed tracing is introduced |

Optional fields are omitted rather than emitted as empty strings or `null` values.

## HTTP request event

Event action: `http.server.request`

Additional fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `event.category` | string array | yes | `["web"]` |
| `event.type` | string array | yes | `["access"]` |
| `event.duration` | integer | yes | Request duration in nanoseconds |
| `http.request.method` | string | yes | HTTP request method, retaining its original casing |
| `http.route` | string | yes | Matched route template, not the raw URL |
| `http.response.status_code` | integer | yes | Response status code |
| `error.type` | string | no | Stable error type for a failed request |
| `error.message` | string | no | Sanitized error message safe for external collection |

`http.route` is an OpenTelemetry semantic-convention attribute without a direct ECS equivalent. It is retained under its OTel name because a route template is more useful and safer than a raw URL path. Elastic can store it as a custom field.

Dynamic API, authentication, and redirect requests are logged. `/health` probes and static frontend requests under `/go` are excluded to avoid repetitive operational noise. HTTP requests use the valid ECS categorization `event.category: ["web"]` and `event.type: ["access"]`. Responses below status 400 use `event.outcome: success`; responses with status 400 or above use `event.outcome: failure`.

Example:

```json
{"@timestamp":"2026-09-20T14:30:00.123456789Z","ecs.version":"9.5.0","log.level":"INFO","message":"HTTP request completed","event.action":"http.server.request","event.category":["web"],"event.type":["access"],"event.outcome":"success","event.duration":4270000,"service.name":"go-url-api","service.version":"1.1.0","service.environment":"production","http.request.id":"871de03da4de4fa68a6d1926f7067037","http.request.method":"GET","http.route":"/:key","http.response.status_code":307}
```

For unexpected server errors, `error.message` is set to a generic value. Internal database errors, stack traces, and request data are not included. An ingestion pipeline may map `error.message` to the OTel `exception.message` attribute where the event represents an exception.

## URL query event

Event action: `go_url.query`

Additional fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `event.category` | string array | yes | `["web"]` |
| `event.type` | string array | yes | `["access"]` |
| `event.duration` | integer | yes | Total HTTP request duration in nanoseconds |
| `http.response.status_code` | integer | yes | Actual HTTP response status code for the request |
| `go_url.query.key` | string | yes | Normalized URL key that was queried |
| `go_url.query.successful` | boolean | yes | Whether the key resolved successfully |
| `source.ip` | string | yes | Validated client IP from the direct connection or trusted forwarded chain |
| `user.name` | string | no | Authenticated username; omitted when no authenticated session is available |

Example:

```json
{"@timestamp":"2026-09-20T14:30:00.122000000Z","ecs.version":"9.5.0","log.level":"INFO","message":"URL query resolved","event.action":"go_url.query","event.category":["web"],"event.type":["access"],"event.outcome":"success","event.duration":4270000,"service.name":"go-url-api","service.version":"1.1.0","service.environment":"production","http.request.id":"871de03da4de4fa68a6d1926f7067037","http.response.status_code":307,"go_url.query.key":"handbook","go_url.query.successful":true,"source.ip":"203.0.113.25","user.name":"alice"}
```

The API emits one event for each distinct, normalized key explicitly requested by the user. URL query events use the valid ECS categorization `event.category: ["web"]` and `event.type: ["access"]`. Both fields are arrays as required by ECS. Multi-key requests can therefore produce both successful and unresolved events. Aliases expanded during resolution do not produce additional query events.

An unresolved query uses the same schema with `event.outcome` set to `failure`, `go_url.query.successful` set to `false`, and the message set to `URL query unresolved`. `http.response.status_code` is the actual response code (for example, 307 for a browser redirect or 404 for a JSON request); `event.duration` is the total request duration. Multi-key requests emit one query event per distinct key, so all events from one request share its request ID, response status, duration, and source IP. A failed key resolution is an expected application outcome and remains at `INFO` severity.

`source.ip` uses the direct peer address from `RemoteAddr` by default. `X-Forwarded-For` is considered only when `ALLOW_FORWARDED_FOR=true`; the configured `FORWARDED_FOR_TRUST_LEVEL` selects the address from the trusted proxy chain. Enable this only when the application is reachable through proxies that overwrite or correctly append the header. Without an authenticated session—including requests from IP allowlisted users who bypass login—`user.name` is omitted.

## Audit events

Audit events are distinct from request and query events: they use `event.dataset: "go_url.audit"`, `event.kind: "event"`, and `event.action` values prefixed with `audit.`. They are always emitted as JSON to stdout, including when `JSON_LOGS=false`.

| `event.action` | `event.type` | Description |
| --- | --- | --- |
| `audit.authentication.login` | `["user", "start"]` | A user successfully authenticated |
| `audit.authentication.logout` | `["user", "end"]` | A local session was successfully cleared |
| `audit.authentication.failure` | `["user", "error"]` | Authentication failed |
| `audit.url.create` | `["creation"]` | A URL was created or a create attempt failed |
| `audit.url.update` | `["change"]` | A URL was updated or an update attempt failed |
| `audit.url.delete` | `["deletion"]` | A URL was deleted or a delete attempt failed |

Audit events include `@timestamp`, `event.outcome`, `http.request.id`, and `source.ip` when available. Authenticated actions also include `user.name`, `user.id` (the provider subject), and `auth.provider`. Failed authentication includes only the provider and a stable `error.type`; unverified usernames are not recorded. URL management events include the normalized `go_url.url.key` and never include the destination URL or request body. Invalid keys are omitted.

`source.ip` uses the direct peer address by default. `X-Forwarded-For` is considered only when `ALLOW_FORWARDED_FOR=true`, using `FORWARDED_FOR_TRUST_LEVEL`; enable this only behind proxies that overwrite or correctly append the header. Authentication failures include invalid login requests, invalid state or nonce, token exchange and validation failures, missing identity claims, and session failures. Credentials, authorization codes, state, nonce, PKCE verifiers, tokens, and provider response bodies are never included.

Provider initialization continues to use the ordinary `authentication` event action; it is not an audit event.

Example:

```json
{"@timestamp":"2026-11-24T20:00:00.123456789Z","ecs.version":"9.5.0","log.level":"INFO","message":"URL updated","event.action":"audit.url.update","event.kind":"event","event.dataset":"go_url.audit","event.category":["web"],"event.type":["change"],"event.outcome":"success","service.name":"go-url-api","http.request.id":"871de03da4de4fa68a6d1926f7067037","source.ip":"203.0.113.25","auth.provider":"ldap","user.name":"alice","user.id":"uid=alice,ou=people,dc=example,dc=com","go_url.url.key":"handbook"}
```

## Application lifecycle events

| `event.action` | Description | Event-specific fields |
| --- | --- | --- |
| `application.startup` | Application started or a startup component failed | `application.component`, `server.port`, `error.type` |
| `application.shutdown` | Shutdown began, completed, or failed | `error.type` on failure |
| `application.panic` | A request panic was recovered | `http.request.id`, `error.type` |

Startup component values identify configuration, database, routes, the HTTP server, or optional Sentry initialization. A Sentry initialization failure is logged but does not prevent the application from starting. Fatal configuration, database, route, and listener failures are logged before the process exits. Panic values and stack traces are not logged; the corresponding HTTP request event contains a sanitized 500 error.

## Database events

| `event.action` | Description | Event-specific fields |
| --- | --- | --- |
| `database.schema` | Database schema initialization succeeded or a migration step failed | `database.system`, `database.operation`, `error.type` on failure |
| `database.connection` | Cleanup after failed schema initialization failed | `database.system`, `database.operation`, `error.type` |

Database error text and connection details are not emitted. `database.operation` identifies the schema step so operators can diagnose failures without exposing credentials or query values.

## Slack events

| `event.action` | Description | Event-specific fields |
| --- | --- | --- |
| `slack.command` | A slash command was accepted or rejected | `http.request.id`, `slack.operation`, `error.type` on failure |
| `slack.bot` | Bot startup, connection, or message operation | `slack.operation`, `error.type` on failure |

## Background task events

`background.task` records the start and stop of the Slack bot event processor and failures while asynchronously recording query metrics or resolving URLs for Slack messages. These events include the stable `background.task` identifier and `error.type` when an operation fails. Raw database and Slack SDK errors are omitted.

## Frontend events

The frontend may submit one event at a time to `POST /api/frontend-events`. The endpoint accepts only version 1 JSON payloads up to 1 KiB and returns `202 Accepted` when logged. Requests are limited to 20 per direct client IP per minute per API process; forwarded IP headers are ignored. The route inherits the application's existing authentication and IP restrictions.

Supported payloads are deliberately narrow:

| Type | Allowed values | Fields |
| --- | --- | --- |
| `error` | `component`: `home`, `search`, `history`, `metrics`, `most_wanted`, `edit_url`, `delete_url`; `code`: `api_request_failed` | `version: 1`, `type`, `component`, `code` |
| `operation` | `operation`: `url.create`, `url.update`, `url.delete`; `outcome`: `success` or `failure` | `version: 1`, `type`, `operation`, `outcome` |

Unknown fields and unsupported values are rejected. The server chooses the event action, severity, outcome for errors, message, and canonical timestamp. It also derives the component for operations, request ID, direct source IP, and (when authenticated) username and auth provider from the server-side session. Client timestamps, severity, identity, IP, URLs, request data, credentials, cookies, tokens, and arbitrary metadata are not accepted or logged.

Frontend reporting is best-effort: a rejected report request is ignored by the browser and is not retried, so logging failures cannot disrupt the UI or create retry loops. The in-memory per-process rate limit resets on restart and is not shared across replicas. For deployments needing a cluster-wide limit, apply a corresponding limit at the ingress.

## Privacy and retention

Query and audit events can associate URL keys or management actions with a source IP and authenticated username or subject. Source IPs and identities may be personal data, and URL keys may reveal user interests or behavior. Collect and retain these fields only for a documented operational need, restrict log access, and use a retention period of 30 days or shorter where practical; follow any stricter organizational or legal policy. Apply retention at the log collector or storage backend. Container log rotation limits local disk use but does not delete copies already shipped to centralized storage.

## Error conventions

Errors use the ECS `error.type` and `error.message` fields:

- `error.type` is a stable category suitable for grouping, not a raw Go error string.
- `error.message` contains only a sanitized message safe for external collection.
- Expected 4xx responses use their public response message where safe.
- Unexpected 5xx responses use `Internal server error`.
- Stack traces are not written to request logs.
- Detailed internal errors can continue to be reported through Sentry.

`error.type` has the same name and meaning in ECS and OTel. ECS `error.message` can be mapped to OTel `exception.message` for exception events.

## Request correlation

- The API accepts `X-Request-ID` only when it is between 1 and 128 characters and contains ASCII letters, digits, `.`, `_`, or `-`.
- Missing or invalid values are replaced with a cryptographically random 32-character hexadecimal ID.
- The canonical value is returned in the `X-Request-ID` response header.
- HTTP and URL-query events from the same request use the same `http.request.id`.
- `trace.id` and `span.id` are reserved for future W3C Trace Context support and are not populated initially.

## Data excluded from logs

The application does not log:

- Raw request URLs or query strings
- Request or response bodies
- Destination URLs or resolved aliases
- Cookies or session identifiers
- Authorization headers
- OAuth codes, access tokens, ID tokens, or client secrets
- Database credentials or connection strings
- Slack tokens or signing secrets
- Sentry DSNs
- Arbitrary request headers
- Internal exception or database error text in HTTP request records

## Configuration

The existing `JSON_LOGS` setting controls output format for all application events:

- `JSON_LOGS=true`: emit one ECS-shaped JSON object per stdout line using this schema.
- `JSON_LOGS=false`: emit the same events in concise human-readable form.

Service metadata configuration:

| Environment variable | Default | Output field | Purpose |
| --- | --- | --- | --- |
| `SERVICE_VERSION` | image build version | `service.version` | Identifies the released application version |
| `SERVICE_ENVIRONMENT` | unset | `service.environment` | Identifies the deployment environment |

`service.name` is fixed as `go-url-api`. Container images set `SERVICE_VERSION` from the image build version, while standalone deployments may set it explicitly.

## Collector mapping summary

An OpenTelemetry Collector or Elastic ingest pipeline should apply these mappings:

| ECS-shaped input | OpenTelemetry output |
| --- | --- |
| `@timestamp` | `Timestamp` |
| collector ingestion time | `ObservedTimestamp` |
| `log.level` | `SeverityText` and derived `SeverityNumber` |
| `message` | `Body` |
| `event.action` | `EventName` |
| `service.name` | resource attribute `service.name` |
| `service.version` | resource attribute `service.version` |
| `service.environment` | resource attribute `deployment.environment.name` |
| `trace.id` | `TraceId` |
| `span.id` | `SpanId` |
| Remaining fields | log record attributes |
