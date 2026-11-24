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

func TestValidateLDAPAuthenticationConfiguration(t *testing.T) {
	validLDAP := LDAP{
		URL:               "ldaps://ldap.example.com:636",
		BindDN:            "cn=service,dc=example,dc=com",
		BindPassword:      "secret",
		UserSearchBase:    "ou=people,dc=example,dc=com",
		UserFilter:        "(&(objectClass=person)(uid={username}))",
		UsernameAttribute: "uid",
	}
	validAuth := func(ldap LDAP) Auth {
		return Auth{
			Enabled:                true,
			Provider:               "ldap",
			SessionToken:           "session",
			MaxAge:                 3600,
			ForwardedForTrustLevel: 1,
			LDAP:                   ldap,
		}
	}
	tables := []struct {
		name      string
		auth      Auth
		appURI    string
		errorText string
	}{
		{name: "valid LDAPS", auth: validAuth(validLDAP), appURI: "https://go.example.com"},
		{name: "valid StartTLS", auth: validAuth(LDAP{
			URL:               "ldap://ldap.example.com:389",
			StartTLS:          true,
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        validLDAP.UserFilter,
			UsernameAttribute: validLDAP.UsernameAttribute,
		}), appURI: "https://go.example.com"},
		{name: "valid local HTTP app", auth: validAuth(validLDAP), appURI: "http://localhost:1323"},
		{name: "missing bind password", auth: validAuth(LDAP{
			URL:               validLDAP.URL,
			BindDN:            validLDAP.BindDN,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        validLDAP.UserFilter,
			UsernameAttribute: validLDAP.UsernameAttribute,
		}), appURI: "https://go.example.com", errorText: "LDAP_BIND_PASSWORD"},
		{name: "LDAP without StartTLS", auth: validAuth(LDAP{
			URL:               "ldap://ldap.example.com",
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        validLDAP.UserFilter,
			UsernameAttribute: validLDAP.UsernameAttribute,
		}), appURI: "https://go.example.com"},
		{name: "StartTLS with LDAPS", auth: validAuth(LDAP{
			URL:               validLDAP.URL,
			StartTLS:          true,
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        validLDAP.UserFilter,
			UsernameAttribute: validLDAP.UsernameAttribute,
		}), appURI: "https://go.example.com", errorText: "cannot be enabled"},
		{name: "invalid username filter placeholder", auth: validAuth(LDAP{
			URL:               validLDAP.URL,
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        "(uid=%s)",
			UsernameAttribute: validLDAP.UsernameAttribute,
		}), appURI: "https://go.example.com", errorText: "{username}"},
		{name: "invalid username attribute", auth: validAuth(LDAP{
			URL:               validLDAP.URL,
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        "(uid={username})",
			UsernameAttribute: "uid)(objectClass=*",
		}), appURI: "https://go.example.com", errorText: "LDAP_USERNAME_ATTRIBUTE"},
		{name: "invalid group filter", auth: validAuth(LDAP{
			URL:               validLDAP.URL,
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        validLDAP.UserFilter,
			UsernameAttribute: validLDAP.UsernameAttribute,
			GroupSearchBase:   "ou=groups,dc=example,dc=com",
			GroupFilter:       "(member={username})",
			AllowedGroups:     []string{"cn=go-url,ou=groups,dc=example,dc=com"},
		}), appURI: "https://go.example.com", errorText: "{user_dn}"},
		{name: "invalid allowed group DN", auth: validAuth(LDAP{
			URL:               validLDAP.URL,
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        validLDAP.UserFilter,
			UsernameAttribute: validLDAP.UsernameAttribute,
			GroupSearchBase:   "ou=groups,dc=example,dc=com",
			GroupFilter:       "(member={user_dn})",
			AllowedGroups:     []string{"not-a-distinguished-name"},
		}), appURI: "https://go.example.com", errorText: "invalid distinguished name"},
		{name: "groups require a search filter", auth: validAuth(LDAP{
			URL:               validLDAP.URL,
			BindDN:            validLDAP.BindDN,
			BindPassword:      validLDAP.BindPassword,
			UserSearchBase:    validLDAP.UserSearchBase,
			UserFilter:        validLDAP.UserFilter,
			UsernameAttribute: validLDAP.UsernameAttribute,
			AllowedGroups:     []string{"cn=go-url,ou=groups,dc=example,dc=com"},
		}), appURI: "https://go.example.com", errorText: "LDAP_GROUP_SEARCH_BASE"},
		{name: "remote app must use HTTPS", auth: validAuth(validLDAP), appURI: "http://go.example.com", errorText: "APP_URI must use HTTPS"},
		{name: "conflicting provider configuration", auth: func() Auth {
			auth := validAuth(validLDAP)
			auth.OIDC = OIDC{
				IssuerURL:     "https://id.example.com/realms/go-url",
				ClientID:      "go-url",
				ClientSecret:  "secret",
				RedirectURL:   "https://go.example.com/oidc/callback",
				Scopes:        []string{"openid", "profile", "email"},
				UsernameClaim: "preferred_username",
				EmailClaim:    "email",
			}
			return auth
		}(), appURI: "https://go.example.com", errorText: "conflicts"},
		{name: "LDAP fields while auth is disabled", auth: Auth{LDAP: validLDAP, ForwardedForTrustLevel: 1}, appURI: "https://go.example.com", errorText: "disabled"},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			err := Validate(Config{Auth: table.auth, AppURI: table.appURI})
			if table.errorText == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if table.errorText != "" && (err == nil || !strings.Contains(err.Error(), table.errorText)) {
				t.Fatalf("error = %v, want text %q", err, table.errorText)
			}
		})
	}
}
