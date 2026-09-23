package config

import (
	"strings"
	"testing"
)

func TestValidateAuthenticationConfiguration(t *testing.T) {
	validOIDC := OIDC{
		IssuerURL:     "https://id.example.com/realms/go-url",
		ClientID:      "go-url",
		ClientSecret:  "secret",
		RedirectURL:   "https://go.example.com/oidc/callback",
		Scopes:        []string{"openid", "profile", "email"},
		UsernameClaim: "preferred_username",
		EmailClaim:    "email",
	}
	tables := []struct {
		name      string
		auth      Auth
		errorText string
	}{
		{name: "disabled", auth: Auth{ForwardedForTrustLevel: 1}},
		{name: "valid OIDC", auth: Auth{Enabled: true, Provider: "oidc", SessionToken: "session", ForwardedForTrustLevel: 1, OIDC: validOIDC}},
		{name: "valid loopback HTTP", auth: Auth{Enabled: true, Provider: "oidc", SessionToken: "session", ForwardedForTrustLevel: 1, OIDC: OIDC{IssuerURL: "http://127.0.0.1:8080", ClientID: "go-url", ClientSecret: "secret", RedirectURL: "http://localhost:1323/oidc/callback", Scopes: []string{"openid"}, UsernameClaim: "sub"}}},
		{name: "missing provider", auth: Auth{Enabled: true, SessionToken: "session", ForwardedForTrustLevel: 1}, errorText: "AUTH_PROVIDER"},
		{name: "missing session token", auth: Auth{Enabled: true, Provider: "oidc", ForwardedForTrustLevel: 1, OIDC: validOIDC}, errorText: "SESSION_TOKEN"},
		{name: "invalid session expiry", auth: Auth{Enabled: true, Provider: "oidc", SessionToken: "session", MaxAge: -1, ForwardedForTrustLevel: 1, OIDC: validOIDC}, errorText: "AUTH_EXPIRY_SECONDS"},
		{name: "inactive provider fields", auth: Auth{Enabled: true, Provider: "oidc", SessionToken: "session", ForwardedForTrustLevel: 1, OIDC: validOIDC, OktaIssuer: "https://example.okta.com"}, errorText: "conflicts"},
		{name: "provider while disabled", auth: Auth{Provider: "oidc", ForwardedForTrustLevel: 1, OIDC: validOIDC}, errorText: "disabled"},
		{name: "missing openid scope", auth: Auth{Enabled: true, Provider: "oidc", SessionToken: "session", ForwardedForTrustLevel: 1, OIDC: OIDC{IssuerURL: "https://id.example.com", ClientID: "go-url", ClientSecret: "secret", RedirectURL: "https://go.example.com/oidc/callback", Scopes: []string{"profile"}, UsernameClaim: "sub"}}, errorText: "openid"},
		{name: "insecure remote issuer", auth: Auth{Enabled: true, Provider: "oidc", SessionToken: "session", ForwardedForTrustLevel: 1, OIDC: OIDC{IssuerURL: "http://id.example.com", ClientID: "go-url", ClientSecret: "secret", RedirectURL: "https://go.example.com/oidc/callback", Scopes: []string{"openid"}, UsernameClaim: "sub"}}, errorText: "HTTPS"},
		{name: "invalid callback path", auth: Auth{Enabled: true, Provider: "oidc", SessionToken: "session", ForwardedForTrustLevel: 1, OIDC: OIDC{IssuerURL: "https://id.example.com", ClientID: "go-url", ClientSecret: "secret", RedirectURL: "https://go.example.com/callback", Scopes: []string{"openid"}, UsernameClaim: "sub"}}, errorText: "/oidc/callback"},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			auth := table.auth
			if auth.Enabled && auth.MaxAge == 0 {
				auth.MaxAge = 3600
			}
			err := Validate(Config{Auth: auth})
			if table.errorText == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if table.errorText != "" && (err == nil || !strings.Contains(err.Error(), table.errorText)) {
				t.Fatalf("error = %v, want text %q", err, table.errorText)
			}
		})
	}
}
