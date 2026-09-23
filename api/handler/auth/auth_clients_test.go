package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
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
