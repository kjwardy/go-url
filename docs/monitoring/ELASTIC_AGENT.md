# Collecting Logs with Elastic Agent

The API writes one ECS-shaped JSON object per stdout line when `JSON_LOGS=true`.
This standalone Elastic Agent policy uses the `filestream` input's NDJSON parser
to read those lines, preserve the ECS and `go_url.*` fields, and send them to
Elasticsearch.

Set `SERVICE_VERSION` and `SERVICE_ENVIRONMENT` on the app container to include
release and deployment metadata in the events. This example reads Kubernetes
container log files from the host; update `paths` to match your runtime and
ensure the agent can access that directory.

```yaml
version: 2

outputs:
  default:
    type: elasticsearch
    hosts:
      - https://elasticsearch.example:9200
    api_key: "${ES_API_KEY}"

inputs:
  - type: filestream
    id: go-url-stdout
    use_output: default
    streams:
      - id: go-url-api
        data_stream:
          type: logs
          dataset: go_url.application
          namespace: default
        paths:
          - /var/log/containers/*go-url*.log
        parsers:
          - container:
              stream: stdout
          - ndjson:
              target: ""
              overwrite_keys: true
              expand_keys: true
```

Set `ES_API_KEY` in the agent's environment to an Elasticsearch API key in
`id:api_key` format; do not put credentials in the policy file. The NDJSON
parser reads one JSON object per line. An empty `target` places fields at the
event root, and `expand_keys` converts dotted ECS names such as `service.name`
and `go_url.query.key` into nested fields. The container parser removes the
runtime envelope before NDJSON parsing. The JSON event's `@timestamp`, ECS
fields, and custom `go_url.*` attributes are retained for Elasticsearch.

For a file that already contains raw application stdout lines, use that path and
remove the `container` parser. If collecting Docker logs, use a runtime-supported
collection path; the Docker `local` logging driver's internal storage is not a
plain-text log file to tail.

For Fleet-managed agents, configure the equivalent `filestream` input and
Elasticsearch output in the agent policy. Input configuration options are
documented in the [Elastic Agent policy reference](https://github.com/elastic/elastic-agent/blob/main/docs/agent-policy.md)
and [Filebeat filestream documentation](https://www.elastic.co/docs/reference/beats/filebeat/filebeat-input-filestream).
