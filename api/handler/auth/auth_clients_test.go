package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

func TestSessionLifecycle(t *testing.T) {
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)), []byte(strings.Repeat("b", 32)))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	manager := newSessionManager(store, 3600, false)
	e := echo.New()

	beginReq := httptest.NewRequest(http.MethodGet, "/private", nil)
	beginRec := httptest.NewRecorder()
	beginContext := e.NewContext(beginReq, beginRec)
	if err := manager.BeginLogin(beginContext, &PendingLogin{Provider: "oidc", State: "state", RedirectPath: "https://evil.example.com"}); err != nil {
		t.Fatal(err)
	}

	callbackReq := httptest.NewRequest(http.MethodGet, "/oidc/callback", nil)
	copyCookies(callbackReq, beginRec.Result().Cookies())
	callbackRec := httptest.NewRecorder()
	callbackContext := e.NewContext(callbackReq, callbackRec)
	pending, err := manager.ConsumePendingLogin(callbackContext, "oidc")
	if err != nil {
		t.Fatal(err)
	}
	redirect, err := manager.CompleteLogin(callbackContext, &Identity{Provider: "oidc", Subject: "subject", Username: "user name", Email: "user@example.com"}, "id-token-secret", pending.RedirectPath)
	if err != nil {
		t.Fatal(err)
	}
	if redirect != "/" {
		t.Errorf("unsafe redirect = %q, want /", redirect)
	}
	userCookie := responseCookie(callbackRec.Result().Cookies(), "user")
	if userCookie == nil || userCookie.Value != "user%20name" || userCookie.HttpOnly || userCookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("user cookie = %#v", userCookie)
	}

	authenticatedReq := httptest.NewRequest(http.MethodGet, "/private", nil)
	copyCookies(authenticatedReq, callbackRec.Result().Cookies())
	authenticatedRec := httptest.NewRecorder()
	authenticatedContext := e.NewContext(authenticatedReq, authenticatedRec)
	if !manager.IsAuthenticated(authenticatedContext) {
		t.Fatal("completed session is not authenticated")
	}
	identity, err := manager.Identity(authenticatedContext)
	if err != nil || identity.Username != "user name" {
		t.Fatalf("identity = %#v, error = %v", identity, err)
	}
	idToken, err := manager.Logout(authenticatedContext)
	if err != nil {
		t.Fatal(err)
	}
	if idToken != "id-token-secret" {
		t.Errorf("logout ID token = %q", idToken)
	}
	clearedUser := responseCookie(authenticatedRec.Result().Cookies(), "user")
	if clearedUser != nil {
		t.Errorf("expected only an expired user cookie, got %#v", clearedUser)
	}
}

func TestPendingLoginIsSingleUse(t *testing.T) {
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)))
	manager := newSessionManager(store, 3600, false)
	e := echo.New()
	beginReq := httptest.NewRequest(http.MethodGet, "/", nil)
	beginRec := httptest.NewRecorder()
	if err := manager.BeginLogin(e.NewContext(beginReq, beginRec), &PendingLogin{Provider: "oidc", State: "state", RedirectPath: "/"}); err != nil {
		t.Fatal(err)
	}
	callbackReq := httptest.NewRequest(http.MethodGet, "/oidc/callback", nil)
	copyCookies(callbackReq, beginRec.Result().Cookies())
	callbackRec := httptest.NewRecorder()
	context := e.NewContext(callbackReq, callbackRec)
	if _, err := manager.ConsumePendingLogin(context, "oidc"); err != nil {
		t.Fatal(err)
	}
	replayReq := httptest.NewRequest(http.MethodGet, "/oidc/callback", nil)
	copyCookies(replayReq, callbackRec.Result().Cookies())
	if _, err := manager.ConsumePendingLogin(e.NewContext(replayReq, httptest.NewRecorder()), "oidc"); err == nil {
		t.Fatal("pending login was accepted more than once")
	}
}

func TestLoginAndLogoutEmitAuditEvents(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(logging.Config{Output: &output})
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	manager := newSessionManager(store, 3600, false)
	manager.logger = logger

	e := echo.New()
	e.Use(logger.Middleware())
	e.GET("/complete", func(c echo.Context) error {
		_, err := manager.CompleteLogin(c, &Identity{
			Provider: "ldap",
			Subject:  "uid=alice,ou=people",
			Username: "alice",
		}, "sensitive-id-token", "/")
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})
	e.POST("/logout", func(c echo.Context) error {
		if _, err := manager.Logout(c); err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	loginRequest := httptest.NewRequest(http.MethodGet, "/complete", nil)
	loginRequest.RemoteAddr = "198.51.100.10:4321"
	loginRequest.Header.Set(echo.HeaderXRequestID, "login-request")
	loginResponse := httptest.NewRecorder()
	e.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login response status = %d, want %d", loginResponse.Code, http.StatusOK)
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
	logoutRequest.RemoteAddr = "198.51.100.10:4321"
	logoutRequest.Header.Set(echo.HeaderXRequestID, "logout-request")
	copyCookies(logoutRequest, loginResponse.Result().Cookies())
	logoutResponse := httptest.NewRecorder()
	e.ServeHTTP(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusOK {
		t.Fatalf("logout response status = %d, want %d", logoutResponse.Code, http.StatusOK)
	}

	var auditEntries []map[string]interface{}
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry["event.dataset"] == "go_url.audit" {
			auditEntries = append(auditEntries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(auditEntries) != 2 {
		t.Fatalf("got %d audit events, want 2; logs: %s", len(auditEntries), output.String())
	}

	wantActions := []string{"audit.authentication.login", "audit.authentication.logout"}
	wantRequestIDs := []string{"login-request", "logout-request"}
	for i, entry := range auditEntries {
		if entry["event.action"] != wantActions[i] || entry["event.outcome"] != "success" {
			t.Errorf("audit event %d = %#v", i, entry)
		}
		if entry["http.request.id"] != wantRequestIDs[i] {
			t.Errorf("audit event %d request id = %#v, want %q", i, entry["http.request.id"], wantRequestIDs[i])
		}
		if entry["source.ip"] != "198.51.100.10" || entry["auth.provider"] != "ldap" || entry["user.name"] != "alice" || entry["user.id"] != "uid=alice,ou=people" {
			t.Errorf("audit event %d identity or source fields = %#v", i, entry)
		}
	}
	if strings.Contains(output.String(), "sensitive-id-token") {
		t.Errorf("audit output leaked ID token: %s", output.String())
	}
}
