# Generic OpenID Connect authentication

`go-url` supports generic OpenID Connect providers that implement discovery and the Authorization Code flow. The implementation uses state, nonce, and PKCE S256 validation and validates ID-token signatures, issuer, audience, and expiry.

## Provider configuration

Register `go-url` as a confidential web application in the identity provider with this callback URL:

```text
https://go.example.com/oidc/callback
```

Configure the application using its environment variables:

```text
ENABLE_AUTH=true
AUTH_PROVIDER=oidc
SESSION_TOKEN=<random session signing key>
SECURE_COOKIES=true

OIDC_ISSUER_URL=https://id.example.com/realms/go-url
OIDC_CLIENT_ID=go-url
OIDC_CLIENT_SECRET=<client secret>
OIDC_REDIRECT_URL=https://go.example.com/oidc/callback
OIDC_SCOPES=openid,profile,email
OIDC_USERNAME_CLAIM=preferred_username
OIDC_EMAIL_CLAIM=email
```

The issuer URL is used for OpenID Provider discovery. It must exactly match the issuer in ID tokens.

`OIDC_SCOPES`, `OIDC_USERNAME_CLAIM`, and `OIDC_EMAIL_CLAIM` default to the values shown above. The scopes must include `openid`. A login is rejected if the selected username claim is missing or is not a non-empty string. The email claim is optional in the returned token.

Only the selected provider may have provider-specific configuration present. `go-url` rejects conflicting Azure AD, Okta, and OIDC settings during startup.

## Local development

Issuer and redirect URLs must use HTTPS. Plain HTTP is accepted only for loopback development addresses such as `localhost`, `127.0.0.1`, and `[::1]`:

```text
OIDC_ISSUER_URL=http://localhost:8081/realms/go-url
OIDC_REDIRECT_URL=http://localhost:1323/oidc/callback
APP_URI=http://localhost:1323
SECURE_COOKIES=false
```

Certificate validation is always enabled for HTTPS discovery, token, and signing-key requests. Install private certificate authorities in the container or host trust store rather than disabling verification.

## Sessions and cookies

Successful authentication creates a server-side session containing the provider name, stable subject, username, and email when available. The session cookie is HTTP-only, uses SameSite Lax, and follows `SECURE_COOKIES` and `AUTH_EXPIRY_SECONDS`.

A separate `user` cookie contains the username for display in the frontend. It is not used for authorization and is cleared during logout.

## Logout

The frontend submits `POST /logout`. `go-url` clears the local session and user cookie first. If discovery advertises an `end_session_endpoint`, the user is redirected to it with an ID-token hint and `APP_URI` as the post-logout redirect. Otherwise the user is redirected directly to `APP_URI`.

## Troubleshooting

- **Discovery failed:** Verify that `OIDC_ISSUER_URL` is reachable from the `go-url` container and that its TLS certificate is trusted.
- **Invalid callback state:** Ensure cookies survive the provider redirect and that all application instances share `SESSION_TOKEN` and session storage.
- **Token exchange failed:** Verify the client ID, client secret, callback URL, and Authorization Code grant configuration.
- **ID token validation failed:** Verify issuer and audience configuration, system clock accuracy, provider signing keys, and token expiry.
- **Nonce validation failed:** Ensure the callback uses the same browser session that initiated login.
- **Authenticated identity has no username:** Configure `OIDC_USERNAME_CLAIM` to a non-empty string claim included in the ID token.
- **Repeated login redirects:** Verify `SECURE_COOKIES`, HTTPS termination, cookie persistence, and shared session storage.

Authentication failures are logged using stable error categories. Authorization codes, state, nonce, PKCE verifiers, tokens, client secrets, and token responses are not logged.
