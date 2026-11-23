# Deploying go-url with Microsoft Entra ID

This Docker Compose example runs go-url and PostgreSQL on a single host. It binds go-url to host loopback for an HTTPS reverse proxy running on the same host. For Entra app registration and identity details, see the [Microsoft Entra authentication guide](../authentication/Azure.md).

See the [authentication overview](../authentication/README.md) for shared provider, session, cookie, and trusted-network guidance.

## 1. Register the Entra application

Create a Web app registration and configure its exact redirect URI:

```text
https://go.example.com/callback
```

Add the Microsoft Graph delegated `User.Read` permission, create a client secret, and record the tenant ID, application client ID, and secret **value**. Assign users and consent according to your organization's policy. See the [provider guide](../authentication/Azure.md) for tenant selection, account permissions, profile mapping, and secret rotation.

## 2. Configure and start Compose

Copy the example environment file, replace the domain, tenant and client IDs, and secret value, then generate unique database and session secrets with `openssl rand -hex 32`. Set `APP_URI` to the public HTTPS origin and `GO_URL_HOSTS` to its host name. If a secret contains `$`, quote it as a single-quoted literal in the Compose environment file. Protect the file:

```sh
cp .env.azure.example .env.azure
chmod 600 .env.azure
```

From the repository directory, pull and start the services:

```sh
docker compose --env-file .env.azure -f docker-compose.auth.yml pull
docker compose --env-file .env.azure -f docker-compose.auth.yml up -d
docker compose --env-file .env.azure -f docker-compose.auth.yml ps
```

Configure the reverse proxy to terminate HTTPS for `APP_URI` and forward to `http://127.0.0.1:1323`, preserving the public host. Do not expose the app directly over public HTTP. If the proxy runs in another container, use a private shared Docker network and proxy to the `app` service on port `1323` instead of publishing it on loopback.

Visit `APP_URI` and confirm sign-in returns to the application. Keep `SESSION_TOKEN` stable across restarts. This example runs one application instance; do not scale it across replicas without configuring shared session storage. Back up the `postgres-data` volume using your normal database backup procedure.

## Operations

- View application logs with `docker compose --env-file .env.azure -f docker-compose.auth.yml logs -f app`.
- If startup reports conflicting provider settings, ensure only the Entra fields are populated and `AUTH_PROVIDER=azure`.
- For callback or consent errors, verify the registered `/callback` URI, Graph `User.Read` permission, tenant ID, and current client secret.
- Docker Compose injects credentials as container environment variables. Protect `.env.azure` and access to the Docker host; use orchestrator secrets where stronger isolation is required.
