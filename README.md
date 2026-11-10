# Go URL

[![CircleCI](https://dl.circleci.com/status-badge/img/circleci/PTuWgk5XLi4dqPB1MLTwdb/3aA338iwpp64iPjk344dZR/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/circleci/PTuWgk5XLi4dqPB1MLTwdb/3aA338iwpp64iPjk344dZR/tree/master)
[![GitHub Actions](https://github.com/kjwardy/go-url/actions/workflows/validate.yml/badge.svg?branch=master)](https://github.com/kjwardy/go-url/actions/workflows/validate.yml?query=branch%3Amaster)

A simple URL shortener written in Go with a React frontend and Postgres database.

# Features

- Shorten URLs using user-defined keys, including keys containing spaces
- Alias a key to another short URL or to multiple keys
- Open multiple pages at once by separating keys with a comma
- Use variables in URLs
- Add, edit, and delete links from the frontend with immediate UI updates
- Search for existing links and view the most popular URLs
- Track timestamped query history, including successful and unresolved requests
- Review the most frequently requested unresolved URLs and create links for them
- View all-time and seven-day query metrics with success rates and a daily histogram
- Persistent light and dark modes
- OpenSearch integration providing suggestions directly in the browser
- Optional authentication using Azure AD, Okta, or generic OpenID Connect
- Slack `/` command integration
- Slackbot integration

# Getting Started

The recommended way to test and deploy is using Docker. You will need to run both the `go-url` app, and the Postgres database.

**Start Postgres**

```sh
docker run -d -P --name db --log-driver=json-file --log-opt max-size=10m --log-opt max-file=3 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=go -e POSTGRES_ADDR=db:5432 postgres:17
```

**Start App**

```sh
docker run -p 1323:1323 --log-driver=json-file --log-opt max-size=10m --log-opt max-file=3 -e HOSTS=localhost -e APP_URI=http://localhost:1323 -e JSON_LOGS=true --link db kjwardy/go-url
```

Alteratively use the docker-compose file and run:

```sh
docker-compose up
```

## Container Log Rotation

The app and Postgres write logs to container stdout/stderr. Docker's default `json-file` driver does not rotate logs unless limits are configured, so logs can otherwise consume the host's available storage. The Compose example limits each service's logs to three files of 10 MB each, a reasonable development default.

Docker supports both `json-file` and `local` logging drivers. Configure a driver per service in Compose, or set a daemon-wide default in the Docker host's `daemon.json` and restart Docker for the change to take effect. For example, use the following for `json-file`:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

For production Docker hosts, prefer the `local` driver, which rotates logs by default and uses a more storage-efficient format. Its defaults are 20 MB per file and five files; to set those limits explicitly:

```json
{
  "log-driver": "local",
  "log-opts": {
    "max-size": "20m",
    "max-file": "5"
  }
}
```

Choose `json-file` instead when your log collector needs to read Docker's JSON log files, and configure rotation as above. Driver changes apply to newly created containers, so recreate existing containers after changing the daemon default. Per-container limits bound each container's local logs; use an external collector and retention policy if logs must be retained beyond that.

On Kubernetes, configure node-level kubelet log rotation (for example, `containerLogMaxSize` and `containerLogMaxFiles`) through your cluster or node configuration; managed Kubernetes providers may expose these settings differently. Deploy a cluster-supported log collector, such as Fluent Bit or the OpenTelemetry Collector, to ship container logs off-node and apply centralized retention. Node rotation limits local storage, while external collection provides durable access to logs.

## Environment Configuration

| Env Var                     | Required | Default        | Example                                        | Description                                                                                            |
| --------------------------- | -------- | -------------- | ---------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `HOSTS`                     | yes      |                | go.domain.com,go2.domain.com                   | List of comma separated hosts that the server will be able to be accessed from                         |
| `BLOCKED_HOSTS`             |          |                | go.domain.com,go2.domain.com                   | List of hosts you want to block from being linked - HOSTS are already included to stop recursive calls |
| `APP_URI`                   | yes      |                | https://go.domain.com                          | Default URI of app - used to link back to app                                                          |
| `PORT`                      |          | 1323           |                                                | Port the app will run on                                                                               |
| `DEBUG`                     |          | false          |                                                | Enable more logging                                                                                    |
| `JSON_LOGS`                 |          | false          |                                                | Emit ECS-shaped API request and URL query logs as JSON to stdout                                        |
| `SERVICE_VERSION`           |          | image version  | v1.1.0                                         | Application version included in structured logs                                                        |
| `SERVICE_ENVIRONMENT`       |          |                | production                                     | Deployment environment included in structured logs                                                     |
| `POSTGRES_ADDR`             |          | localhost:5432 |                                                | Postgres db address                                                                                    |
| `POSTGRES_DATABASE`         |          | go             |                                                | Postgres db name                                                                                       |
| `POSTGRES_USER`             |          | postgres       |                                                | Postgres user                                                                                          |
| `POSTGRES_PASS`             |          | password       |                                                | Postgres password                                                                                      |
| `SLACK_TOKEN`               |          |                | xoxb-xxxxxxxxx-xxxxxxxx-xxxx                   | Slack OAuth token to enable slackbot                                                                   |
| `SENTRY_API_DSN`            |          |                |                                                | Sentry DSN for go API                                                                                  |
| `SENTRY_FRONTEND_DSN`       |          |                |                                                | Sentry DSN for react frontend                                                                          |
| `SLACK_SIGNING_SECRET`      |          |                | xxxxxxxxxxx                                    | Slack signing secret to enable Slack `/go` command                                                     |
| `SLACK_TEAM_ID`             |          |                | Txxxxxxxx                                      | Slack team id to restrict slash command responses to single team                                       |

### Authentication Configuration

| Env Var                     | Required | Default              | Example                                        | Description                                                                                           |
| --------------------------- | -------- | -------------------- | ---------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `ENABLE_AUTH`               |          | false                |                                                | Enable authentication                                                                                 |
| `AUTH_PROVIDER`             | auth     |                      | oidc                                           | Active provider: `azure`, `okta`, or `oidc`                                                           |
| `AUTH_EXPIRY_SECONDS`       |          | 2592000              |                                                | Auth cookie expiry (default 30 days)                                                                  |
| `SECURE_COOKIES`            |          | true                 |                                                | Use secure HTTPS-only cookies                                                                         |
| `SESSION_TOKEN`             | auth     |                      |                                                | Secret signing key for application sessions                                                           |
| `AD_TENANT_ID`              | Azure   |                      |                                                | Azure AD tenant ID                                                                                    |
| `AD_CLIENT_ID`              | Azure   |                      |                                                | Azure AD client ID                                                                                    |
| `AD_CLIENT_SECRET`          | Azure   |                      |                                                | Azure AD client secret                                                                                |
| `OKTA_CLIENT_ID`            | Okta    |                      |                                                | Okta client ID                                                                                        |
| `OKTA_CLIENT_SECRET`        | Okta    |                      |                                                | Okta client secret                                                                                    |
| `OKTA_ISSUER`               | Okta    |                      | https://dev-123.oktapreview.com/oauth2/default | Okta issuer URL                                                                                       |
| `OIDC_ISSUER_URL`           | OIDC    |                      | https://id.example.com/realms/go-url           | OpenID Provider issuer used for discovery                                                             |
| `OIDC_CLIENT_ID`            | OIDC    |                      | go-url                                         | OIDC client ID                                                                                        |
| `OIDC_CLIENT_SECRET`        | OIDC    |                      |                                                | OIDC client secret                                                                                    |
| `OIDC_REDIRECT_URL`         | OIDC    |                      | https://go.example.com/oidc/callback           | Registered OIDC callback URL                                                                          |
| `OIDC_SCOPES`               |          | openid,profile,email |                                                | Comma-separated OIDC scopes                                                                           |
| `OIDC_USERNAME_CLAIM`       |          | preferred_username   |                                                | ID-token claim used as the username                                                                   |
| `OIDC_EMAIL_CLAIM`          |          | email                |                                                | ID-token claim used as the email address                                                              |
| `ALLOWED_IPS`               |          |                      | 110.1.10.2,1.1.22.0/24                         | IP addresses or CIDRs that bypass authentication                                                      |
| `ALLOW_FORWARDED_FOR`       |          | false                |                                                | Retrieve origin IP from X-Forwarded-For. Enable only behind a trusted proxy                            |
| `FORWARDED_FOR_TRUST_LEVEL` |          | 1                    |                                                | Number of trusted proxy levels in X-Forwarded-For                                                     |

See the [generic OIDC guide](docs/authentication/OIDC.md) for provider setup, security behavior, logout, and troubleshooting.

See the [API logging schema](docs/monitoring/LOGGING_SCHEMA.md) for event fields, ECS and OpenTelemetry mappings, correlation behavior, and excluded sensitive data.
