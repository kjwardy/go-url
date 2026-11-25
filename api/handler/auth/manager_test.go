package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-pg/pg"
	"github.com/gorilla/sessions"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

type logoutTestProvider struct {
	idToken string
}

func (p *logoutTestProvider) Name() string                  { return "oidc" }
func (p *logoutTestProvider) RegisterRoutes(*echo.Echo)     {}
func (p *logoutTestProvider) BeginLogin(echo.Context) error { return nil }
func (p *logoutTestProvider) CallbackPath() string          { return "/oidc/callback" }
func (p *logoutTestProvider) LogoutURL(idToken string) (string, bool) {
	p.idToken = idToken
	return "https://id.example.com/logout", true
}

func TestManagerLogoutClearsSessionBeforeProviderRedirect(t *testing.T) {
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)), []byte(strings.Repeat("b", 32)))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	sessionManager := newSessionManager(store, 3600, false)
	provider := &logoutTestProvider{}
	manager := &Manager{provider: provider, sessions: sessionManager, appURI: "https://go.example.com"}
	e := echo.New()
	manager.RegisterRoutes(e)

	beginReq := httptest.NewRequest(http.MethodGet, "/", nil)
	beginRec := httptest.NewRecorder()
	beginContext := e.NewContext(beginReq, beginRec)
	if err := sessionManager.BeginLogin(beginContext, &PendingLogin{Provider: "oidc", RedirectPath: "/"}); err != nil {
		t.Fatal(err)
	}

	loginReq := httptest.NewRequest(http.MethodGet, "/oidc/callback", nil)
	copyCookies(loginReq, beginRec.Result().Cookies())
	loginRec := httptest.NewRecorder()
	loginContext := e.NewContext(loginReq, loginRec)
	if _, err := sessionManager.CompleteLogin(loginContext, &Identity{Provider: "oidc", Subject: "subject", Username: "user"}, "id-token", "/"); err != nil {
		t.Fatal(err)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", nil)
	copyCookies(logoutReq, loginRec.Result().Cookies())
	logoutRec := httptest.NewRecorder()
	e.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusSeeOther || logoutRec.Header().Get("Location") != "https://id.example.com/logout" {
		t.Fatalf("logout response = %d %q", logoutRec.Code, logoutRec.Header().Get("Location"))
	}
	if provider.idToken != "id-token" {
		t.Errorf("provider ID token hint = %q", provider.idToken)
	}
	if responseCookie(logoutRec.Result().Cookies(), "user") != nil {
		t.Error("logout did not expire the user cookie")
	}
}

func TestManagerIdentityReturnsAuthenticatedSessionIdentity(t *testing.T) {
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)), []byte(strings.Repeat("b", 32)))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	sessionManager := newSessionManager(store, 3600, false)
	manager := &Manager{sessions: sessionManager}
	e := echo.New()

	loginReq := httptest.NewRequest(http.MethodGet, "/", nil)
	loginRec := httptest.NewRecorder()
	if _, err := sessionManager.CompleteLogin(
		e.NewContext(loginReq, loginRec),
		&Identity{Provider: "oidc", Subject: "subject", Username: "user"},
		"",
		"/",
	); err != nil {
		t.Fatal(err)
	}

	identityReq := httptest.NewRequest(http.MethodPost, "/api/frontend-events", nil)
	copyCookies(identityReq, loginRec.Result().Cookies())
	identity, err := manager.Identity(e.NewContext(identityReq, httptest.NewRecorder()))
	if err != nil {
		t.Fatal(err)
	}
	if identity.Provider != "oidc" || identity.Username != "user" {
		t.Errorf("identity = %#v", identity)
	}
}

func TestLDAPManagerExposesLoginRoutesWithoutSession(t *testing.T) {
	ldapConfig := config.LDAP{
		URL:               "ldaps://ldap.example.com",
		BindDN:            "cn=service,dc=example,dc=com",
		BindPassword:      "secret",
		UserSearchBase:    "ou=users,dc=example,dc=com",
		UserFilter:        "(uid={username})",
		UsernameAttribute: "uid",
	}
	manager, err := NewManager(context.Background(), config.Auth{
		Enabled:                true,
		Provider:               "ldap",
		SessionToken:           strings.Repeat("s", 32),
		MaxAge:                 3600,
		SecureCookies:          true,
		ForwardedForTrustLevel: 1,
		LDAP:                   ldapConfig,
	}, "https://go.example.com", logging.New(logging.Config{}), &pg.DB{})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		ldapLoginPath,
		ldapLoginPage,
		"/go/config.js",
		"/go/manifest.json",
		"/go/favicon.ico",
		"/go/logo.png",
		"/go/static/js/main.abc123.js",
		"/go/static/css/main.abc123.css",
	} {
		if !manager.IsPublic(path) {
			t.Errorf("LDAP login asset %q must be public so an unauthenticated user can sign in", path)
		}
	}
	if manager.IsPublic("/api/search") {
		t.Error("ordinary API routes must still require authentication")
	}
	if manager.IsPublic("/go/private") {
		t.Error("unlisted application routes must still require authentication")
	}
}
