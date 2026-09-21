# Proposed API Logging Schema

This document proposes the JSON schema for API logs written to stdout. It is intended for review before implementation.

The application will emit Elastic Common Schema (ECS)-shaped JSON because Elastic Agent is the expected initial consumer. The selected fields either match OpenTelemetry semantic conventions directly or have a documented mapping to the OpenTelemetry Logs Data Model. Fields specific to Go URL use the `go_url.*` namespace.

ECS and OpenTelemetry semantic conventions are converging, but they are not identical. The application will not emit duplicate ECS and OpenTelemetry representations of the same value. Translation to OTLP should happen in the OpenTelemetry Collector or Elastic ingest pipeline.

## ECS and OpenTelemetry mapping

| JSON field | ECS status | OpenTelemetry destination | Description |
| --- | --- | --- | --- |
| `@timestamp` | native | `Timestamp` | Time the event occurred, in RFC 3339 format with nanosecond precision |
| `ecs.version` | native | record attribute | ECS version followed by the event; currently `9.5.0` |
| `log.level` | native | `SeverityText` | Text severity: `DEBUG`, `INFO`, `WARN`, or `ERROR` |
| `message` | native | `Body` | Short, stable human-readable description |
| `event.action` | native | `EventName` | Stable machine-readable event type |
| `event.outcome` | native | record attribute | Whether the operation succeeded or failed |
| `event.duration` | native | record attribute | Event duration in nanoseconds |
| `service.name` | exact match | resource attribute | Service emitting the record |
| `service.version` | exact match | resource attribute | Released application version |
| `service.environment` | native | `deployment.environment.name` resource attribute | Deployment environment |
| `http.request.id` | native | record attribute | Request correlation identifier |
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
| `trace.id` | string | no | W3C trace ID when distributed tracing is introduced |
| `span.id` | string | no | W3C span ID when distributed tracing is introduced |

Optional fields are omitted rather than emitted as empty strings or `null` values.

## HTTP request event

Event action: `http.server.request`

Additional fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `event.duration` | integer | yes | Request duration in nanoseconds |
| `http.request.method` | string | yes | HTTP request method, retaining its original casing |
| `http.route` | string | yes | Matched route template, not the raw URL |
| `http.response.status_code` | integer | yes | Response status code |
| `error.type` | string | no | Stable error type for a failed request |
| `error.message` | string | no | Sanitized error message safe for external collection |

`http.route` is an OpenTelemetry semantic-convention attribute without a direct ECS equivalent. It is retained under its OTel name because a route template is more useful and safer than a raw URL path. Elastic can store it as a custom field.

Dynamic API, authentication, and redirect requests are logged. `/health` probes and static frontend requests under `/go` are excluded to avoid repetitive operational noise. Responses below status 400 use `event.outcome: success`; responses with status 400 or above use `event.outcome: failure`.

Example:

```json
{"@timestamp":"2026-09-20T14:30:00.123456789Z","ecs.version":"9.5.0","log.level":"INFO","message":"HTTP request completed","event.action":"http.server.request","event.outcome":"success","event.duration":4270000,"service.name":"go-url-api","service.version":"1.1.0","service.environment":"production","http.request.id":"871de03da4de4fa68a6d1926f7067037","http.request.method":"GET","http.route":"/:key","http.response.status_code":307}
```

For unexpected server errors, `error.message` is set to a generic value. Internal database errors, stack traces, and request data are not included. An ingestion pipeline may map `error.message` to the OTel `exception.message` attribute where the event represents an exception.

## URL query event

Event action: `go_url.query`

Additional fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `go_url.query.key` | string | yes | Normalized URL key that was queried |
| `go_url.query.successful` | boolean | yes | Whether the key resolved successfully |

Example:

```json
{"@timestamp":"2026-09-20T14:30:00.122000000Z","ecs.version":"9.5.0","log.level":"INFO","message":"URL query resolved","event.action":"go_url.query","event.outcome":"success","service.name":"go-url-api","service.version":"1.1.0","service.environment":"production","http.request.id":"871de03da4de4fa68a6d1926f7067037","go_url.query.key":"handbook","go_url.query.successful":true}
```

The API emits one event for each distinct, normalized key explicitly requested by the user. Multi-key requests can therefore produce both successful and unresolved events. Aliases expanded during resolution do not produce additional query events.

An unresolved query uses the same schema with `event.outcome` set to `failure`, `go_url.query.successful` set to `false`, and the message set to `URL query unresolved`. A failed resolution is an expected application outcome and remains at `INFO` severity.

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

The existing `JSON_LOGS` setting controls output format:

- `JSON_LOGS=true`: emit one ECS-shaped JSON object per stdout line using this schema.
- `JSON_LOGS=false`: retain concise human-readable request and query logs.

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
