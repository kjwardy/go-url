# Deploying go-url with Okta

This Docker Compose example runs go-url and PostgreSQL on a single host. It binds go-url to host loopback for an HTTPS reverse proxy running on the same host. For Okta application setup, issuer selection, and identity mapping, see the [Okta authentication guide](../authentication/Okta.md).

See the [authentication overview](../authentication/README.md) for shared provider, session, cookie, and trusted-network guidance.

## 1. Register the Okta application

Create an OIDC Web Application using the Authorization Code grant, assign the users or groups who may sign in, and register this exact sign-in redirect URI:

```text
https://go.example.com/okta/callback
```

Record the client ID, client secret, and issuer for the authorization server used by the app. The example uses `https://<yourOktaDomain>/oauth2/default`; configure that authorization server to issue the `openid`, `profile`, and `email` scopes and an ID-token username claim. See the [provider guide](../authentication/Okta.md) for issuer selection, claims, assignment policies, and secret rotation.

## 2. Configure and start Compose

Copy the example environment file, replace the domain, Okta client values, and issuer, then generate unique database and session secrets with `openssl rand -hex 32`. Set `APP_URI` to the public HTTPS origin and `GO_URL_HOSTS` to its host name. If a secret contains `$`, quote it as a single-quoted literal in the Compose environment file. Protect the file:

```sh
cp .env.okta.example .env.okta
chmod 600 .env.okta
```

From the repository directory, pull and start the services:

```sh
docker compose --env-file .env.okta -f docker-compose.auth.yml pull
docker compose --env-file .env.okta -f docker-compose.auth.yml up -d
docker compose --env-file .env.okta -f docker-compose.auth.yml ps
```

Configure the reverse proxy to terminate HTTPS for `APP_URI` and forward to `http://127.0.0.1:1323`, preserving the public host. Do not expose the app directly over public HTTP. If the proxy runs in another container, use a private shared Docker network and proxy to the `app` service on port `1323` instead of publishing it on loopback.

Visit `APP_URI` and confirm sign-in returns to the application. Keep `SESSION_TOKEN` stable across restarts. This example runs one application instance; do not scale it across replicas without configuring shared session storage. Back up the `postgres-data` volume using your normal database backup procedure.

## Operations

- View application logs with `docker compose --env-file .env.okta -f docker-compose.auth.yml logs -f app`.
- If startup reports conflicting provider settings, ensure only the Okta fields are populated and `AUTH_PROVIDER=okta`.
- For callback or token errors, verify the registered `/okta/callback` URI, exact issuer, client ID, user assignment, claims, and current client secret.
- Docker Compose injects credentials as container environment variables. Protect `.env.okta` and access to the Docker host; use orchestrator secrets where stronger isolation is required.
