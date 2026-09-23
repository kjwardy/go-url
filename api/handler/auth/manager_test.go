package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
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
