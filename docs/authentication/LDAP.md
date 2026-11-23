# LDAP authentication

`go-url` can authenticate users against an LDAP directory using a service-account search followed by a user bind. Authentication creates the same server-side session used by the other providers.

For a complete Docker Compose deployment, including a sample environment file and HTTPS reverse-proxy requirements, see the [LDAP deployment guide](../deployment/LDAP.md).

## Configuration

Configure one provider only. Example using LDAPS:

```text
ENABLE_AUTH=true
AUTH_PROVIDER=ldap
SESSION_TOKEN=<random session signing key>
SECURE_COOKIES=true
APP_URI=https://go.example.com

LDAP_URL=ldaps://ldap.example.com:636
LDAP_BIND_DN=cn=go-url,ou=services,dc=example,dc=com
LDAP_BIND_PASSWORD=<service account password>
LDAP_USER_SEARCH_BASE=ou=people,dc=example,dc=com
LDAP_USER_FILTER=(&(objectClass=person)(uid={username}))
LDAP_USERNAME_ATTRIBUTE=uid
```

`LDAP_USERNAME_ATTRIBUTE` defaults to `uid`. If `LDAP_USER_FILTER` is omitted, it defaults to `(uid={username})` or uses the configured username attribute. The user filter must contain exactly one `{username}` placeholder; submitted usernames are LDAP-filter escaped before substitution. The directory search must return exactly one user entry.

Use either `ldaps://` for implicit TLS or `ldap://` with `LDAP_STARTTLS=true`. Plain LDAP without StartTLS is rejected. TLS certificate verification is enabled by default, TLS 1.2 or newer is required, and `LDAP_CA_CERT_FILE` can add a PEM-encoded CA certificate to the system trust store.

**Do not enable `LDAP_INSECURE_SKIP_VERIFY` in production.** It disables LDAP server certificate verification and makes the connection vulnerable to interception. Use it only in a controlled development environment when no trusted certificate is available.

## Restricting access by group

By default, any directory user with valid credentials can sign in. To limit access, configure all three group settings:

```text
LDAP_GROUP_SEARCH_BASE=ou=groups,dc=example,dc=com
LDAP_GROUP_FILTER=(&(objectClass=groupOfNames)(member={user_dn}))
LDAP_ALLOWED_GROUPS=cn=go-url,ou=groups,dc=example,dc=com;cn=admins,ou=groups,dc=example,dc=com
```

The group filter must contain exactly one `{user_dn}` placeholder. Each returned group entry's distinguished name is compared with the semicolon-separated allowed DNs. Adjust the filter and search base to match the directory's schema and membership model. If the user is not in an allowed group, sign-in is denied.

## Login and sessions

Unauthenticated requests are redirected to `/go/login`; after a successful login, the user returns to the originally requested local path. Opening the login page directly signs in to the application root. The React form posts credentials as JSON to the LDAP login endpoint. Successful authentication creates the shared server-side session and frontend username cookie; the session cookie is HTTP-only and follows `SECURE_COOKIES` and `AUTH_EXPIRY_SECONDS`.

Remote deployments must use an HTTPS `APP_URI`. HTTP is accepted for loopback development addresses such as `localhost`, `127.0.0.1`, and `[::1]`; set `SECURE_COOKIES=false` only for local HTTP development.

## Troubleshooting

- **Authentication service unavailable:** Check that the LDAP host is reachable from the `go-url` container, the bind account can search the configured base, and the TLS certificate chain is trusted.
- **Valid credentials are denied:** Verify the user search filter and username attribute. If group restrictions are configured, verify the group search base, membership filter, and allowed group DNs.
- **Login redirects repeatedly:** Verify that browser cookies are enabled and that all application instances use the same `SESSION_TOKEN`.

Authentication failures use stable log categories. Passwords, bind errors, user attributes, and directory response details are not logged.
