# Deploying go-url with LDAP authentication

This example runs the published `go-url` image and PostgreSQL with Docker Compose. It expects a DNS name for the application, a reachable LDAP server, and an HTTPS reverse proxy on the Docker host. Replace the example domain, directory names, and credentials with values for your environment.

## 1. Prepare configuration

Copy the example environment file and generate unique secrets:

```sh
cp .env.ldap.example .env.ldap
openssl rand -hex 32
```

Set `POSTGRES_PASS` to a generated database password and `SESSION_TOKEN` to a different generated value. Fill in `LDAP_URL`, the service-account DN and password, the user search base, and the filter/attribute that match your directory. Set `GO_URL_HOSTS` to the host name users will visit and `APP_URI` to its HTTPS URL. Protect the file because it contains passwords:

```sh
chmod 600 .env.ldap
```

The sample restricts sign-in to the configured group. Change the group search base, filter, and allowed group DN to match your directory. To allow every user found by the user search, remove all three `LDAP_GROUP_*`/`LDAP_ALLOWED_GROUPS` settings from `.env.ldap`.

The sample uses LDAPS and system-trusted certificates. For `ldap://` with StartTLS, set `LDAP_URL=ldap://ldap.example.com:389` and `LDAP_STARTTLS=true`. Never disable certificate verification in production. For a private CA, add its certificate to the application container:

```yaml
    volumes:
      - ./certs/ldap-ca.pem:/etc/go-url/ldap-ca.pem:ro
```

Then uncomment `LDAP_CA_CERT_FILE` in `.env.ldap` and set it to `/etc/go-url/ldap-ca.pem`. Place the certificate at `./certs/ldap-ca.pem` next to the Compose file.

## 2. Configure HTTPS

The Compose file binds the application to `127.0.0.1:1323`, suitable for a reverse proxy running on the Docker host. Configure that proxy to:

- terminate HTTPS for the `APP_URI` host name and forward requests to `http://127.0.0.1:1323`;
- preserve the request host and support the application's normal HTTP methods;
- avoid caching login pages, login API responses, or session cookies.

Do not expose the app directly over public HTTP: login credentials are submitted to the application and must be protected in transit. If the proxy runs in another container, use a shared Docker network and proxy to the `app` service on port `1323` instead of publishing the app on loopback. Keep that network private.

## 3. Start the services

Run these commands from the repository directory, where `docker-compose.ldap.yml` and `.env.ldap` are located:

```sh
docker compose --env-file .env.ldap -f docker-compose.ldap.yml pull
docker compose --env-file .env.ldap -f docker-compose.ldap.yml up -d
docker compose --env-file .env.ldap -f docker-compose.ldap.yml ps
```

Visit `APP_URI`. An unauthenticated request should lead to the LDAP sign-in page. Keep `SESSION_TOKEN` stable across restarts so existing sessions remain valid. This example runs one application instance; do not scale it across replicas without configuring shared session storage. Back up the `postgres-data` volume using your normal database backup procedure.

## Operations and troubleshooting

- View service logs with `docker compose --env-file .env.ldap -f docker-compose.ldap.yml logs -f app`.
- If startup reports invalid authentication configuration, check that `AUTH_PROVIDER=ldap`, all required LDAP settings are present, and no settings for another authentication provider are configured.
- If login reports the authentication service is unavailable, confirm the container can resolve and reach `LDAP_URL`, the service account can search `LDAP_USER_SEARCH_BASE`, and the TLS chain is trusted.
- If credentials work but access is denied, check the user filter and username attribute. With group restrictions, confirm the membership filter and allowed group DN match the directory's schema.
- If the app redirects repeatedly to login, verify HTTPS termination, `APP_URI`, and that browser cookies are enabled.

Docker Compose passes credentials as container environment variables. Restrict access to the Docker host and `.env.ldap`; for deployments requiring stronger secret isolation, use your orchestrator's secret-management facilities.
