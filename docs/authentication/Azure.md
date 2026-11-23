# Microsoft Entra ID (Azure AD) authentication

See the [authentication overview](README.md) for provider comparison, shared sessions and cookies, reverse-proxy requirements, and trusted-network configuration.

The Azure provider uses the Microsoft identity platform authorization-code flow and Microsoft Graph to retrieve the signed-in user's profile. Configure one Entra tenant and register the callback URL below.

## Register the application

In the Microsoft Entra admin center, create an App registration for go-url:

1. Select the account types appropriate for your organization. `AD_TENANT_ID` determines the tenant-specific sign-in authority; use the tenant that owns the registration.
2. Under **Authentication**, add a **Web** platform redirect URI:

   ```text
   https://go.example.com/callback
   ```

   Replace the host with the public `APP_URI` host. The callback path is always `/callback`; it is not separately configurable.
3. Under **API permissions**, add Microsoft Graph **Delegated** permission `User.Read`. Grant consent according to your organization's consent policy.
4. Create a client secret under **Certificates & secrets**. Store the value securely; Microsoft displays it only when created. Set a rotation reminder and replace `AD_CLIENT_SECRET` before the old secret expires.

The provider requests the fixed `User.Read` scope and calls `https://graph.microsoft.com/v1.0/me`. It uses the returned `id` as the stable subject, `mail` as the email and username when present, and otherwise `userPrincipalName` as the username. `displayName` is not used for login identity.

## Configure go-url

Set the common session and provider settings in the application environment:

```text
ENABLE_AUTH=true
AUTH_PROVIDER=azure
SESSION_TOKEN=<random session signing key>
AUTH_EXPIRY_SECONDS=2592000
SECURE_COOKIES=true
APP_URI=https://go.example.com

AD_TENANT_ID=<Microsoft Entra tenant ID>
AD_CLIENT_ID=<Application (client) ID>
AD_CLIENT_SECRET=<client secret value>
```

`APP_URI` is the public HTTPS origin with no trailing slash. The callback is derived as `${APP_URI}/callback`. Azure scopes, claims, and callback path are fixed by the implementation.

## Sessions and logout

The session stores the provider name, Entra user ID, username, and email when available. The HTTP-only session cookie follows `AUTH_EXPIRY_SECONDS` and `SECURE_COOKIES`; the separate frontend `user` cookie is display-only. `POST /logout` clears the local go-url session. It does not sign the user out of Microsoft Entra ID, so a later login may reuse the user's Microsoft SSO session.

For HTTPS termination, proxy to the application over a protected internal connection and keep `APP_URI` set to the external HTTPS origin. Forwarded client IPs are not required for authentication; enable `ALLOW_FORWARDED_FOR` only if using `ALLOWED_IPS` behind a trusted proxy that overwrites the forwarding header.

## Troubleshooting

- **Reply URL mismatch:** Confirm the registered URI is exactly `https://<APP_URI-host>/callback`, with the same public host and port.
- **Consent or Graph error:** Check that `User.Read` is a delegated Microsoft Graph permission and that tenant policy allows the required consent.
- **Authentication fails after secret rotation:** Update `AD_CLIENT_SECRET` with the new secret *value*, then restart go-url. Do not use the secret's identifier.
- **User details cannot be retrieved:** Verify the app can reach `login.microsoftonline.com` and `graph.microsoft.com`, and that the signed-in account is permitted by tenant policy.
- **Authenticated identity has no username:** Ensure the Graph profile returns a non-empty `mail` or `userPrincipalName`.
- **Repeated login redirects:** Check HTTPS proxy configuration, `APP_URI`, browser cookie acceptance, and the lifetime set by `AUTH_EXPIRY_SECONDS`.

The provider uses the tenant-specific OAuth endpoints and validates state in the callback. Client secrets, authorization codes, access tokens, and Graph profile responses are not logged.
