# API Logging Schema

This document describes the implemented JSON schema for application logs written to stdout. Structured request, URL query, lifecycle, database, authentication, Slack, and background-task events are available when `JSON_LOGS=true`.

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
| `go_url.query.key` | string | yes | Normalized URL key that was queried |
| `go_url.query.successful` | boolean | yes | Whether the key resolved successfully |

Example:

```json
{"@timestamp":"2026-09-20T14:30:00.122000000Z","ecs.version":"9.5.0","log.level":"INFO","message":"URL query resolved","event.action":"go_url.query","event.category":["web"],"event.type":["access"],"event.outcome":"success","service.name":"go-url-api","service.version":"1.1.0","service.environment":"production","http.request.id":"871de03da4de4fa68a6d1926f7067037","go_url.query.key":"handbook","go_url.query.successful":true}
```

The API emits one event for each distinct, normalized key explicitly requested by the user. URL query events use the valid ECS categorization `event.category: ["web"]` and `event.type: ["access"]`. Both fields are arrays as required by ECS. Multi-key requests can therefore produce both successful and unresolved events. Aliases expanded during resolution do not produce additional query events.

An unresolved query uses the same schema with `event.outcome` set to `failure`, `go_url.query.successful` set to `false`, and the message set to `URL query unresolved`. A failed resolution is an expected application outcome and remains at `INFO` severity.

## Authentication failure event

Event action: `authentication`

Authentication failures use `event.outcome: failure`, `auth.provider`, and a stable `error.type`. Supported categories include invalid state or nonce, token exchange and validation failures, missing identity claims, and session failures. Authorization codes, state, nonce, PKCE verifiers, tokens, client secrets, and provider response bodies are never included.

Example:

```json
{"@timestamp":"2026-09-22T14:30:00.122000000Z","ecs.version":"9.5.0","log.level":"WARN","message":"Authentication failed","event.action":"authentication","event.outcome":"failure","service.name":"go-url-api","http.request.id":"871de03da4de4fa68a6d1926f7067037","auth.provider":"oidc","error.type":"invalid_state"}
```

The same action records provider initialization success and failure with the configured provider. Request authentication failures include `http.request.id`; startup events do not.

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
