# Authentication

`go-url` supports four authentication providers. Choose one based on the identity system you already operate; configure only that provider's settings.

| Method | Best fit | User identity and access |
| --- | --- | --- |
| [Microsoft Entra ID (Azure AD)](Azure.md) | Organizations using Microsoft 365 / Entra ID | Microsoft Graph `id` is the subject; `mail`, or `userPrincipalName` when mail is empty, is the username. Access follows Entra ID sign-in policy. |
| [Okta](Okta.md) | Organizations using Okta | The ID token `sub` is the subject; `preferred_username`, falling back to `email`, is the username. Access follows Okta app assignments and authorization-server policy. |
| [Generic OIDC](OIDC.md) | Any compatible OpenID Connect provider | The ID token `sub` is the subject; configured claims provide username and optional email. |
| [LDAP](LDAP.md) | Existing LDAP directories without an OIDC broker | The user's DN is the subject; the configured directory attribute supplies the username. Optional group search can restrict access. |

## Shared settings and behavior

Enable exactly one provider with `ENABLE_AUTH=true` and `AUTH_PROVIDER=azure`, `okta`, `oidc`, or `ldap`. `SESSION_TOKEN` is required and must be a long, random secret. Provider-specific settings for another provider must be unset; startup rejects conflicting provider configuration. See the [environment variable reference](../../README.md#authentication-configuration).

`AUTH_EXPIRY_SECONDS` controls session lifetime and defaults to 2,592,000 seconds (30 days). Successful login creates an application session; its cookie is HTTP-only, SameSite Lax, and Secure when `SECURE_COOKIES=true` (the default). The separate `user` cookie is only for frontend display and is not an authorization credential. `POST /logout` clears the local session and display cookie. Azure AD, Okta, and LDAP sign-out is local to go-url; the identity provider may retain its own SSO session. Generic OIDC also redirects to the provider's end-session endpoint when discovery advertises one.

When running multiple application instances, use the same `SESSION_TOKEN` and ensure they share the session store.

For remote deployments, set `APP_URI` to the canonical HTTPS origin with no trailing slash and set `SECURE_COOKIES=true`. Place the application behind an HTTPS reverse proxy and register the exact provider callback URL shown in its guide. `SECURE_COOKIES=false` is for loopback-only local development; it is not a workaround for an incorrectly terminated production proxy.

`ALLOW_FORWARDED_FOR` is disabled by default. Enable it only when the application is reachable exclusively through trusted proxies that overwrite `X-Forwarded-For`, and set `FORWARDED_FOR_TRUST_LEVEL` to the number of trusted proxy hops. Do not trust a client-supplied forwarding header. `ALLOWED_IPS` has special behavior: when authentication is enabled, matching clients bypass login; when authentication is disabled, it acts as an allowlist and other clients receive `403 Forbidden`.

## Running without authentication

Leave `ENABLE_AUTH=false` and `AUTH_PROVIDER` plus all provider-specific settings unset. With no `ALLOWED_IPS`, the application does not restrict access by client IP. Use this mode only when a separate, trusted network boundary (for example, a private network or VPN) enforces who can reach the application. To enforce an IP allowlist instead, configure `ALLOWED_IPS`; if using `X-Forwarded-For`, follow the trusted-proxy requirements above.

## Troubleshooting across providers

- **Redirect URI mismatch or callback failure:** Use the exact callback path and public HTTPS origin specified by the provider guide. `APP_URI` must match the registered scheme, host, and port.
- **Repeated redirects or expired session:** Ensure the browser accepts cookies, all instances use the same `SESSION_TOKEN`, and the session has not expired.
- **Unexpected client IP or allowlist behavior:** Check the direct connection path and only enable forwarded-IP handling behind proxies you control.
- **TLS or certificate errors:** Verify the provider and application trust chains. For LDAP, also check the configured CA file and StartTLS/LDAPS mode.
- **Login succeeds at the provider but go-url rejects it:** Check the configured issuer, provider app assignment, required API permissions, and required username claim or directory attribute.

Authentication failures are logged using stable categories. Passwords, client secrets, authorization codes, tokens, and raw provider responses are not logged.
