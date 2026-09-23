package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gorilla/sessions"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

type fakeOIDCServer struct {
	server            *httptest.Server
	key               *rsa.PrivateKey
	signingKey        *rsa.PrivateKey
	mu                sync.Mutex
	nonce             string
	tokenNonce        string
	challenge         string
	audience          string
	issuer            string
	subject           string
	expiry            time.Time
	username          interface{}
	tokenStatus       int
	lastTokenVerifier string
}

func newFakeOIDCServer(t *testing.T) *fakeOIDCServer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeOIDCServer{key: key, signingKey: key, audience: "go-url", subject: "subject-123", expiry: time.Now().Add(time.Hour), username: "kieran"}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.serveHTTP))
	fake.issuer = fake.server.URL
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeOIDCServer) serveHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                                f.server.URL,
			"authorization_endpoint":                f.server.URL + "/authorize",
			"token_endpoint":                        f.server.URL + "/token",
			"jwks_uri":                              f.server.URL + "/keys",
			"end_session_endpoint":                  f.server.URL + "/logout",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	case "/keys":
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &f.key.PublicKey, KeyID: "test-key", Algorithm: "RS256", Use: "sig"}}})
	case "/token":
		f.token(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeOIDCServer) token(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastTokenVerifier = r.Form.Get("code_verifier")
	if f.tokenStatus != 0 {
		http.Error(w, "token exchange rejected", f.tokenStatus)
		return
	}
	if oauth2.S256ChallengeFromVerifier(f.lastTokenVerifier) != f.challenge {
		http.Error(w, "PKCE validation failed", http.StatusBadRequest)
		return
	}
	signer, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: f.signingKey}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key"))
	claims := jwt.Claims{
		Issuer:   f.issuer,
		Subject:  f.subject,
		Audience: jwt.Audience{f.audience},
		Expiry:   jwt.NewNumericDate(f.expiry),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	nonce := f.nonce
	if f.tokenNonce != "" {
		nonce = f.tokenNonce
	}
	privateClaims := map[string]interface{}{"nonce": nonce, "preferred_username": f.username, "email": "kieran@example.com"}
	rawIDToken, _ := jwt.Signed(signer).Claims(claims).Claims(privateClaims).Serialize()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": "access-token-value",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     rawIDToken,
	})
}

func (f *fakeOIDCServer) prepareAuthorization(location string) (string, error) {
	parsed, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	f.mu.Lock()
	f.nonce = query.Get("nonce")
	f.challenge = query.Get("code_challenge")
	f.mu.Unlock()
	if query.Get("code_challenge_method") != "S256" {
		return "", fmt.Errorf("code challenge method = %q", query.Get("code_challenge_method"))
	}
	if query.Get("scope") != "openid profile email" {
		return "", fmt.Errorf("scope = %q", query.Get("scope"))
	}
	return query.Get("state"), nil
}

func TestOIDCAuthorizationCodeFlow(t *testing.T) {
	fake := newFakeOIDCServer(t)
	provider, sessionManager, loggerOutput := newTestOIDCProvider(t, fake)
	e := echo.New()
	e.Use(logging.New(logging.Config{JSON: true, Output: loggerOutput}).Middleware())
	provider.RegisterRoutes(e)
	e.GET("/private", provider.BeginLogin)

	loginReq := httptest.NewRequest(http.MethodGet, "/private?from=test", nil)
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("login status = %d, want %d", loginRec.Code, http.StatusTemporaryRedirect)
	}
	state, err := fake.prepareAuthorization(loginRec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if state == "" || fake.nonce == "" || fake.challenge == "" {
		t.Fatal("authorization redirect is missing state, nonce, or PKCE challenge")
	}

	callbackReq := httptest.NewRequest(http.MethodGet, "/oidc/callback?state="+url.QueryEscape(state)+"&code=valid-code", nil)
	copyCookies(callbackReq, loginRec.Result().Cookies())
	callbackRec := httptest.NewRecorder()
	e.ServeHTTP(callbackRec, callbackReq)
	if callbackRec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("callback status = %d, body = %s", callbackRec.Code, callbackRec.Body.String())
	}
	if callbackRec.Header().Get("Location") != "/private?from=test" {
		t.Errorf("callback location = %q", callbackRec.Header().Get("Location"))
	}
	if fake.lastTokenVerifier == "" {
		t.Error("token request did not include PKCE verifier")
	}
	if user := responseCookie(callbackRec.Result().Cookies(), "user"); user == nil || user.Value != "kieran" {
		t.Fatalf("user cookie = %#v", user)
	}

	identityReq := httptest.NewRequest(http.MethodGet, "/identity", nil)
	copyCookies(identityReq, callbackRec.Result().Cookies())
	identityRec := httptest.NewRecorder()
	identityContext := e.NewContext(identityReq, identityRec)
	identity, err := sessionManager.Identity(identityContext)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Provider != "oidc" || identity.Subject != "subject-123" || identity.Username != "kieran" || identity.Email != "kieran@example.com" {
		t.Errorf("identity = %#v", identity)
	}
	if strings.Contains(loggerOutput.String(), "access-token-value") || strings.Contains(loggerOutput.String(), "valid-code") {
		t.Errorf("sensitive value appeared in logs: %s", loggerOutput.String())
	}
}

func TestOIDCCallbackFailures(t *testing.T) {
	tables := []struct {
		name       string
		configure  func(*fakeOIDCServer)
		state      string
		wantStatus int
		wantError  string
	}{
		{name: "invalid state", state: "invalid", wantStatus: http.StatusBadRequest, wantError: "invalid_state"},
		{name: "token exchange", configure: func(f *fakeOIDCServer) { f.tokenStatus = http.StatusUnauthorized }, wantStatus: http.StatusBadGateway, wantError: "token_exchange_failed"},
		{name: "invalid issuer", configure: func(f *fakeOIDCServer) { f.issuer = "https://another-issuer.example.com" }, wantStatus: http.StatusUnauthorized, wantError: "token_validation_failed"},
		{name: "invalid audience", configure: func(f *fakeOIDCServer) { f.audience = "another-client" }, wantStatus: http.StatusUnauthorized, wantError: "token_validation_failed"},
		{name: "invalid signature", configure: func(f *fakeOIDCServer) { f.signingKey, _ = rsa.GenerateKey(rand.Reader, 2048) }, wantStatus: http.StatusUnauthorized, wantError: "token_validation_failed"},
		{name: "expired token", configure: func(f *fakeOIDCServer) { f.expiry = time.Now().Add(-time.Hour) }, wantStatus: http.StatusUnauthorized, wantError: "token_validation_failed"},
		{name: "invalid nonce", configure: func(f *fakeOIDCServer) { f.tokenNonce = "another-nonce" }, wantStatus: http.StatusUnauthorized, wantError: "invalid_nonce"},
		{name: "missing subject", configure: func(f *fakeOIDCServer) { f.subject = "" }, wantStatus: http.StatusUnauthorized, wantError: "missing_subject"},
		{name: "missing username", configure: func(f *fakeOIDCServer) { f.username = nil }, wantStatus: http.StatusUnauthorized, wantError: "missing_username_claim"},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			fake := newFakeOIDCServer(t)
			if table.configure != nil {
				table.configure(fake)
			}
			provider, _, loggerOutput := newTestOIDCProvider(t, fake)
			e := echo.New()
			e.Use(logging.New(logging.Config{JSON: true, Output: loggerOutput}).Middleware())
			provider.RegisterRoutes(e)
			e.GET("/private", provider.BeginLogin)

			loginReq := httptest.NewRequest(http.MethodGet, "/private", nil)
			loginRec := httptest.NewRecorder()
			e.ServeHTTP(loginRec, loginReq)
			state, err := fake.prepareAuthorization(loginRec.Header().Get("Location"))
			if err != nil {
				t.Fatal(err)
			}
			if table.state != "" {
				state = table.state
			}
			callbackReq := httptest.NewRequest(http.MethodGet, "/oidc/callback?state="+url.QueryEscape(state)+"&code=code-secret", nil)
			copyCookies(callbackReq, loginRec.Result().Cookies())
			callbackRec := httptest.NewRecorder()
			e.ServeHTTP(callbackRec, callbackReq)
			if callbackRec.Code != table.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", callbackRec.Code, table.wantStatus, callbackRec.Body.String())
			}
			if !strings.Contains(loggerOutput.String(), `"error.type":"`+table.wantError+`"`) {
				t.Errorf("missing sanitized error %q in logs: %s", table.wantError, loggerOutput.String())
			}
			for _, sensitive := range []string{"code-secret", "access-token-value", fake.nonce} {
				if sensitive != "" && strings.Contains(loggerOutput.String(), sensitive) {
					t.Errorf("sensitive value appeared in logs: %s", loggerOutput.String())
				}
			}
		})
	}
}

func TestOIDCDiscoveryRejectsUntrustedTLS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"issuer": "https://" + r.Host})
	}))
	defer server.Close()
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)))
	_, err := NewOIDCProvider(context.Background(), config.OIDC{
		IssuerURL:     server.URL,
		ClientID:      "go-url",
		ClientSecret:  "secret",
		RedirectURL:   "https://go.example.com/oidc/callback",
		Scopes:        []string{"openid"},
		UsernameClaim: "sub",
	}, "https://go.example.com", newSessionManager(store, 3600, true), logging.New(logging.Config{}))
	if err == nil {
		t.Fatal("discovery unexpectedly trusted a self-signed certificate")
	}
}

func TestOIDCLogoutURL(t *testing.T) {
	fake := newFakeOIDCServer(t)
	provider, _, _ := newTestOIDCProvider(t, fake)
	location, ok := provider.LogoutURL("id-token")
	if !ok {
		t.Fatal("provider logout was not enabled")
	}
	parsed, _ := url.Parse(location)
	if parsed.Path != "/logout" || parsed.Query().Get("post_logout_redirect_uri") != "http://localhost:1323" || parsed.Query().Get("id_token_hint") != "id-token" {
		t.Errorf("logout URL = %q", location)
	}
}

func newTestOIDCProvider(t *testing.T, fake *fakeOIDCServer) (*OIDCProvider, *SessionManager, *bytes.Buffer) {
	t.Helper()
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)), []byte(strings.Repeat("b", 32)))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	sessionManager := newSessionManager(store, 3600, false)
	var loggerOutput bytes.Buffer
	logger := logging.New(logging.Config{JSON: true, Output: &loggerOutput})
	provider, err := NewOIDCProvider(context.Background(), config.OIDC{
		IssuerURL:     fake.server.URL,
		ClientID:      "go-url",
		ClientSecret:  "client-secret",
		RedirectURL:   "http://localhost:1323/oidc/callback",
		Scopes:        []string{"openid", "profile", "email"},
		UsernameClaim: "preferred_username",
		EmailClaim:    "email",
	}, "http://localhost:1323", sessionManager, logger)
	if err != nil {
		t.Fatal(err)
	}
	return provider, sessionManager, &loggerOutput
}

func copyCookies(req *http.Request, cookies []*http.Cookie) {
	seen := make(map[string]bool)
	for i := len(cookies) - 1; i >= 0; i-- {
		cookie := cookies[i]
		if cookie.MaxAge >= 0 && !seen[cookie.Name] {
			req.AddCookie(cookie)
			seen[cookie.Name] = true
		}
	}
}

func responseCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for i := len(cookies) - 1; i >= 0; i-- {
		if cookies[i].Name == name && cookies[i].MaxAge >= 0 {
			return cookies[i]
		}
	}
	return nil
}
