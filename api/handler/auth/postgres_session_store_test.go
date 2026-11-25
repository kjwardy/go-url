package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/kjwardy/go-url/api/config"
	"github.com/labstack/echo"
)

type memorySessionRecord struct {
	data      []byte
	expiresAt time.Time
}

type memorySessionRepository struct {
	records map[string]memorySessionRecord
}

func newMemorySessionRepository() *memorySessionRepository {
	return &memorySessionRepository{records: make(map[string]memorySessionRecord)}
}

func (r *memorySessionRepository) load(_ context.Context, id string) ([]byte, error) {
	record, ok := r.records[id]
	if !ok || !record.expiresAt.After(time.Now()) {
		return nil, errSessionNotFound
	}
	return append([]byte(nil), record.data...), nil
}

func (r *memorySessionRepository) save(_ context.Context, id string, data []byte, expiresAt time.Time) error {
	r.records[id] = memorySessionRecord{
		data:      append([]byte(nil), data...),
		expiresAt: expiresAt,
	}
	return nil
}

func (r *memorySessionRepository) delete(_ context.Context, id string) error {
	delete(r.records, id)
	return nil
}

func testPostgresSessionManager(repository sessionRepository) *SessionManager {
	authConfig := config.Auth{
		SessionToken:  strings.Repeat("s", 32),
		MaxAge:        3600,
		SecureCookies: false,
	}
	return newSessionManager(newSessionStore(repository, authConfig), authConfig.MaxAge, authConfig.SecureCookies)
}

func TestPostgresSessionStorePersistsSessionsAcrossInstances(t *testing.T) {
	repository := newMemorySessionRepository()
	firstManager := testPostgresSessionManager(repository)
	e := echo.New()

	loginResponse := httptest.NewRecorder()
	_, err := firstManager.CompleteLogin(
		e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), loginResponse),
		&Identity{Provider: "ldap", Subject: "uid=alice,ou=people", Username: "alice"},
		"id-token",
		"/",
	)
	if err != nil {
		t.Fatal(err)
	}
	cookie := responseCookie(loginResponse.Result().Cookies(), sessionName)
	if cookie == nil {
		t.Fatal("login did not set a session cookie")
	}

	secondManager := testPostgresSessionManager(repository)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(cookie)
	identity, err := secondManager.Identity(e.NewContext(request, httptest.NewRecorder()))
	if err != nil {
		t.Fatal(err)
	}
	if identity.Provider != "ldap" || identity.Subject != "uid=alice,ou=people" || identity.Username != "alice" {
		t.Errorf("identity = %#v", identity)
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
	logoutRequest.AddCookie(cookie)
	logoutResponse := httptest.NewRecorder()
	idToken, err := secondManager.Logout(e.NewContext(logoutRequest, logoutResponse))
	if err != nil {
		t.Fatal(err)
	}
	if idToken != "id-token" {
		t.Errorf("logout ID token = %q, want id-token", idToken)
	}
	if len(repository.records) != 0 {
		t.Errorf("session records after logout = %d, want 0", len(repository.records))
	}
	sessionCookieExpired := false
	for _, responseCookie := range logoutResponse.Result().Cookies() {
		if responseCookie.Name == sessionName && responseCookie.MaxAge < 0 {
			sessionCookieExpired = true
		}
	}
	if !sessionCookieExpired {
		t.Error("logout did not expire the session cookie")
	}
}

func TestPostgresSessionStoreStartsFreshLoginWhenSessionRowIsMissing(t *testing.T) {
	firstManager := testPostgresSessionManager(newMemorySessionRepository())
	e := echo.New()
	loginResponse := httptest.NewRecorder()
	_, err := firstManager.CompleteLogin(
		e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), loginResponse),
		&Identity{Provider: "ldap", Subject: "uid=alice,ou=people", Username: "alice"},
		"",
		"/",
	)
	if err != nil {
		t.Fatal(err)
	}
	staleCookie := responseCookie(loginResponse.Result().Cookies(), sessionName)
	if staleCookie == nil {
		t.Fatal("login did not set a session cookie")
	}

	secondManager := testPostgresSessionManager(newMemorySessionRepository())
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(staleCookie)
	response := httptest.NewRecorder()
	context := e.NewContext(request, response)
	if secondManager.IsAuthenticated(context) {
		t.Fatal("missing server-side session was treated as authenticated")
	}
	if err := secondManager.BeginLogin(context, &PendingLogin{
		Provider:     "ldap",
		RedirectPath: "/",
	}); err != nil {
		t.Fatalf("login failed for stale session cookie: %v", err)
	}
	freshCookie := responseCookie(response.Result().Cookies(), sessionName)
	if freshCookie == nil || freshCookie.Value == staleCookie.Value {
		t.Fatal("stale session cookie was not replaced")
	}

	pendingRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	pendingRequest.AddCookie(freshCookie)
	pending, err := secondManager.PendingLogin(e.NewContext(pendingRequest, httptest.NewRecorder()), "ldap")
	if err != nil {
		t.Fatal(err)
	}
	if pending.Provider != "ldap" {
		t.Errorf("pending provider = %q, want ldap", pending.Provider)
	}
}

type failingSessionRepository struct {
	err error
}

func (r failingSessionRepository) load(context.Context, string) ([]byte, error) {
	return nil, r.err
}

func (r failingSessionRepository) save(context.Context, string, []byte, time.Time) error {
	return r.err
}

func (r failingSessionRepository) delete(context.Context, string) error {
	return r.err
}

func TestPostgresSessionStoreReturnsRepositoryFailures(t *testing.T) {
	wantErr := errors.New("database unavailable")
	store := newSessionStore(failingSessionRepository{err: wantErr}, config.Auth{
		SessionToken: strings.Repeat("s", 32),
		MaxAge:       3600,
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	encodedID, err := securecookie.EncodeMulti(sessionName, "session-id", store.codecs...)
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(&http.Cookie{Name: sessionName, Value: encodedID})
	_, err = store.New(request, sessionName)
	if !errors.Is(err, wantErr) {
		t.Fatalf("New error = %v, want %v", err, wantErr)
	}
}
