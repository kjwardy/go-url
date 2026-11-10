# Collecting Logs with the OpenTelemetry Collector

The API writes one ECS-shaped JSON object per stdout line when `JSON_LOGS=true`.
This example uses the OpenTelemetry Collector Contrib `filelog` receiver to
parse those records, maps the fields to the OpenTelemetry Logs Data Model, and
exports to Elasticsearch and the debug console.

Set `SERVICE_VERSION` and `SERVICE_ENVIRONMENT` on the app container to include
release and deployment metadata in the events. This example reads Kubernetes
container log files from the host; update `include` to match your runtime and
ensure the collector can access that directory.

```yaml
receivers:
  filelog/go_url:
    include:
      - /var/log/pods/*/*/*.log
    include_file_path: true
    start_at: end
    operators:
      - type: container
      - type: json_parser
        parse_from: body
        parse_to: attributes
        timestamp:
          parse_from: attributes["@timestamp"]
          layout_type: gotime
          layout: "2006-01-02T15:04:05.999999999Z07:00"
        severity:
          parse_from: attributes["log.level"]
          mapping:
            debug: DEBUG
            info: INFO
            warn: WARN
            error: ERROR
      - type: trace_parser
        trace_id:
          parse_from: attributes["trace.id"]
        span_id:
          parse_from: attributes["span.id"]

processors:
  transform/go_url:
    error_mode: ignore
    log_statements:
      - context: log
        statements:
          - set(log.body, log.attributes["message"])
          - delete_key(log.attributes, "message")
          - set(log.event_name, log.attributes["event.action"])
          - delete_key(log.attributes, "event.action")
          - set(resource.attributes["service.name"], log.attributes["service.name"])
          - delete_key(log.attributes, "service.name")
          - set(resource.attributes["service.version"], log.attributes["service.version"]) where log.attributes["service.version"] != nil
          - delete_key(log.attributes, "service.version") where log.attributes["service.version"] != nil
          - set(resource.attributes["deployment.environment.name"], log.attributes["service.environment"]) where log.attributes["service.environment"] != nil
          - delete_key(log.attributes, "service.environment") where log.attributes["service.environment"] != nil
          - set(resource.attributes["data_stream.type"], "logs")
          - set(resource.attributes["data_stream.dataset"], "go_url")
          - set(resource.attributes["data_stream.namespace"], "default")
      - context: scope
        statements:
          - set(scope.attributes["elastic.mapping.mode"], "otel")

  batch:
    timeout: 5s

exporters:
  elasticsearch:
    endpoint: "${env:ELASTICSEARCH_ENDPOINT}"
    api_key: "${env:ELASTICSEARCH_API_KEY}"
  debug:
    verbosity: detailed

service:
  pipelines:
    logs:
      receivers: [filelog/go_url]
      processors: [transform/go_url, batch]
      exporters: [elasticsearch, debug]
```

Set `ELASTICSEARCH_ENDPOINT` and `ELASTICSEARCH_API_KEY` in the collector
environment. `ELASTICSEARCH_API_KEY` must be an encoded Elasticsearch API key.
The exporter uses OTel mapping mode; the transform processor selects it through
the `elastic.mapping.mode` scope attribute and routes logs to the
`logs-go_url.otel-default` data stream. The detailed debug exporter is useful
for validating transformed records; remove it from the pipeline outside
development if its volume or output visibility is undesirable.

The JSON parser assigns `@timestamp` to the log record `Timestamp`, maps
`log.level` to `SeverityText` and `SeverityNumber` (DEBUG 5, INFO 9, WARN 13,
ERROR 17), and the transform processor maps `message` to `Body` and
`event.action` to `EventName`. `service.name`, `service.version`, and
`service.environment` become resource attributes `service.name`,
`service.version`, and `deployment.environment.name`. `trace.id` and `span.id`
become the native `TraceId` and `SpanId` fields when present. The filelog
receiver sets `ObservedTimestamp`; the source timestamp remains `Timestamp`.
All other parsed fields remain log record attributes, including
`event.category`, `event.type`, and custom `go_url.*` fields.

For a file that already contains raw application stdout lines, use that path
and remove the `container` operator. If collecting Docker logs, use a
runtime-supported collection path; the Docker `local` logging driver's
internal storage is not a plain-text log file to tail. The `container` operator
decodes Docker, CRI-O, and containerd log envelopes before the application JSON
is parsed.

This configuration requires a Collector distribution that includes the
`filelog` receiver, `transform` processor, and Elasticsearch exporter, such as
`otelcol-contrib`. See the upstream [filelog receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver),
[transform processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/transformprocessor),
and [Elasticsearch exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/elasticsearchexporter)
documentation for component options.
