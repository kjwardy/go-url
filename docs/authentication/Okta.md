# Okta authentication

See the [authentication overview](README.md) for provider comparison, shared sessions and cookies, reverse-proxy requirements, and trusted-network configuration.

The Okta provider uses a confidential OIDC web application and the authorization-code flow. It validates the ID token's signature, issuer, audience, and nonce before creating a go-url session.

## Create and configure the Okta application

1. In the Okta Admin Console, create an **OIDC - OpenID Connect** **Web Application** integration using the **Authorization Code** grant.
2. Add this sign-in redirect URI, replacing the host with the public `APP_URI` host:

   ```text
   https://go.example.com/okta/callback
   ```

   The callback path is fixed as `/okta/callback`; `APP_URI` must not have a trailing slash.
3. Assign the application to the people or groups allowed to sign in.
4. Enable the `openid`, `profile`, and `email` scopes for the authorization server. The provider requests these scopes automatically.
5. Copy the client ID and client secret into the go-url configuration. Rotate the secret according to your organization's policy and update go-url before the old secret expires.

Set `OKTA_ISSUER` to the issuer of the authorization server used by this application. For an Okta custom authorization server, this commonly looks like `https://<yourOktaDomain>/oauth2/default`; use the issuer shown in the server's metadata, not an authorize or token endpoint URL. The configured issuer determines the `/v1/authorize` and `/v1/token` endpoints. Use the same authorization server for the issuer and the application sign-in policy.

The verified ID token's `sub` becomes the stable subject. go-url uses `preferred_username` as the username, falling back to `email`; `email` is also stored when available. Configure the authorization server to include a non-empty `preferred_username` or `email` claim in its ID token. The `profile` and `email` scopes are requested, but the provider's claim configuration and user profile data must still supply the values.

## Configure go-url

```text
ENABLE_AUTH=true
AUTH_PROVIDER=okta
SESSION_TOKEN=<random session signing key>
AUTH_EXPIRY_SECONDS=2592000
SECURE_COOKIES=true
APP_URI=https://go.example.com

OKTA_CLIENT_ID=<Okta client ID>
OKTA_CLIENT_SECRET=<Okta client secret>
OKTA_ISSUER=https://<yourOktaDomain>/oauth2/default
```

`APP_URI` is the public HTTPS origin with no trailing slash. Register `${APP_URI}/okta/callback` exactly. The scope list and callback path are fixed by the implementation.

## Sessions and logout

The session contains the issuer-validated `sub`, username, and email when available. Its HTTP-only cookie follows `AUTH_EXPIRY_SECONDS` and `SECURE_COOKIES`; a separate `user` cookie is used only for frontend display. `POST /logout` clears the local go-url session but does not initiate Okta's end-session flow, so Okta may keep the browser's SSO session active.

Terminate HTTPS at a trusted reverse proxy and set `APP_URI` to the external HTTPS origin. `ALLOW_FORWARDED_FOR` is unrelated to the OIDC callback; use it only with an exclusively trusted proxy that overwrites `X-Forwarded-For` and only when `ALLOWED_IPS` needs the client address.

## Troubleshooting

- **Redirect URI mismatch:** Check the exact `${APP_URI}/okta/callback` value in the Okta app and ensure the external host and port match.
- **Invalid issuer or token:** Set `OKTA_ISSUER` to the exact issuer for the selected authorization server; do not append `/v1/authorize` or `/v1/token`.
- **Invalid audience or authorization server:** Confirm the app's client ID matches `OKTA_CLIENT_ID` and that the authorization server issues tokens for that client.
- **Missing username:** Ensure the ID token includes a non-empty `preferred_username` or `email`; adjust the authorization server's claim mapping and scopes.
- **User cannot sign in:** Confirm the user or group is assigned to the Okta application and permitted by its sign-on policy.
- **Repeated login redirects or expired sessions:** Check proxy HTTPS configuration, cookie acceptance, shared `SESSION_TOKEN`, and `AUTH_EXPIRY_SECONDS`.
- **Secret rotation failure:** Update `OKTA_CLIENT_SECRET` with the current secret value and restart go-url.

Okta failures are reported using stable categories; client secrets, authorization codes, ID tokens, and raw token responses are not logged.
