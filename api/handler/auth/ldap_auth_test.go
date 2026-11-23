package auth

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/sessions"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

const (
	testLDAPBindDN      = "cn=service,dc=example,dc=com"
	testLDAPBindSecret  = "service-bind-secret"
	testLDAPUserDN      = "uid=alice,ou=users,dc=example,dc=com"
	testLDAPUserSecret  = "user-password-secret"
	testLDAPGroupDN     = "cn=go-url,ou=groups,dc=example,dc=com"
	testLDAPSearchBase  = "ou=users,dc=example,dc=com"
	testLDAPGroupBase   = "ou=groups,dc=example,dc=com"
	testLDAPUsername    = "alice"
	testLDAPRequestPath = "/api/ldap/login"
)

type fakeLDAPConn struct {
	userEntry      *ldap.Entry
	groupEntries   []*ldap.Entry
	groupSearchErr error
	serviceBindErr error
	userBindErr    error
	searchErr      error
	startTLSErr    error
	bindDNs        []string
	bindPasswords  []string
	searchRequests []*ldap.SearchRequest
	startTLSConfig *tls.Config
	timeout        time.Duration
	closed         bool
}

func (c *fakeLDAPConn) Bind(username, password string) error {
	c.bindDNs = append(c.bindDNs, username)
	c.bindPasswords = append(c.bindPasswords, password)
	if username == testLDAPBindDN {
		return c.serviceBindErr
	}
	return c.userBindErr
}

func (c *fakeLDAPConn) Search(request *ldap.SearchRequest) (*ldap.SearchResult, error) {
	c.searchRequests = append(c.searchRequests, request)
	if request.BaseDN == testLDAPGroupBase {
		return &ldap.SearchResult{Entries: c.groupEntries}, c.groupSearchErr
	}
	return &ldap.SearchResult{Entries: []*ldap.Entry{c.userEntry}}, c.searchErr
}

func (c *fakeLDAPConn) StartTLS(config *tls.Config) error {
	c.startTLSConfig = config
	return c.startTLSErr
}

func (c *fakeLDAPConn) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func (c *fakeLDAPConn) Close() error {
	c.closed = true
	return nil
}

func ldapTestConfig() config.LDAP {
	return config.LDAP{
		URL:               "ldap://ldap.example.com",
		StartTLS:          true,
		BindDN:            testLDAPBindDN,
		BindPassword:      testLDAPBindSecret,
		UserSearchBase:    testLDAPSearchBase,
		UserFilter:        "(uid={username})",
		UsernameAttribute: "uid",
		GroupSearchBase:   testLDAPGroupBase,
		GroupFilter:       "(&(objectClass=groupOfNames)(member={user_dn}))",
		AllowedGroups:     []string{testLDAPGroupDN},
	}
}

func newTestSessionManager() *SessionManager {
	store := sessions.NewCookieStore([]byte(strings.Repeat("a", 32)), []byte(strings.Repeat("b", 32)))
	store.Options = &sessions.Options{Path: "/", MaxAge: 3600, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	return newSessionManager(store, 3600, false)
}

func newTestLDAPProvider(t *testing.T, ldapConfig config.LDAP, conn *fakeLDAPConn, output *bytes.Buffer) (*LDAPProvider, *SessionManager, *echo.Echo) {
	t.Helper()
	sessionManager := newTestSessionManager()
	logger := logging.New(logging.Config{JSON: true, Output: output})
	provider, err := NewLDAPProvider(ldapConfig, sessionManager, logger)
	if err != nil {
		t.Fatal(err)
	}
	provider.dialLDAP = func(string, ...ldap.DialOpt) (ldapConnection, error) {
		return conn, nil
	}
	e := echo.New()
	e.Use(logger.Middleware())
	provider.RegisterRoutes(e)
	e.GET("/private", provider.BeginLogin)
	return provider, sessionManager, e
}

func beginLDAPLogin(t *testing.T, e *echo.Echo) []*http.Cookie {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/private?from=test", nil)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != ldapLoginPage {
		t.Fatalf("login start response = %d %q", response.Code, response.Header().Get("Location"))
	}
	return response.Result().Cookies()
}

func postLDAPCredentials(e *echo.Echo, cookies []*http.Cookie, username, password string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(ldapLoginRequest{Username: username, Password: password})
	request := httptest.NewRequest(http.MethodPost, testLDAPRequestPath, bytes.NewReader(body))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	copyCookies(request, cookies)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	return response
}

func TestLDAPLoginCreatesExistingSession(t *testing.T) {
	conn := &fakeLDAPConn{
		userEntry: ldap.NewEntry(testLDAPUserDN, map[string][]string{
			"uid":  {testLDAPUsername},
			"mail": {"alice@example.com"},
		}),
		groupEntries: []*ldap.Entry{ldap.NewEntry(strings.ToUpper(testLDAPGroupDN), nil)},
	}
	var output bytes.Buffer
	_, sessions, e := newTestLDAPProvider(t, ldapTestConfig(), conn, &output)
	cookies := beginLDAPLogin(t, e)
	response := postLDAPCredentials(e, cookies, testLDAPUsername, testLDAPUserSecret)

	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var result map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["redirect"] != "/private?from=test" {
		t.Errorf("redirect = %q", result["redirect"])
	}
	authenticatedRequest := httptest.NewRequest(http.MethodGet, "/private", nil)
	copyCookies(authenticatedRequest, response.Result().Cookies())
	identity, err := sessions.Identity(e.NewContext(authenticatedRequest, httptest.NewRecorder()))
	if err != nil {
		t.Fatal(err)
	}
	if identity.Provider != "ldap" || identity.Subject != testLDAPUserDN || identity.Username != testLDAPUsername || identity.Email != "alice@example.com" {
		t.Errorf("identity = %#v", identity)
	}
	if len(conn.bindDNs) != 2 || conn.bindDNs[0] != testLDAPBindDN || conn.bindDNs[1] != testLDAPUserDN {
		t.Errorf("bind DNs = %#v", conn.bindDNs)
	}
	if conn.bindPasswords[0] != testLDAPBindSecret || conn.bindPasswords[1] != testLDAPUserSecret {
		t.Errorf("unexpected bind passwords")
	}
	if len(conn.searchRequests) != 2 || conn.searchRequests[0].Filter != "(uid=alice)" || !strings.Contains(conn.searchRequests[1].Filter, testLDAPUserDN) {
		t.Errorf("search filters = %#v", conn.searchRequests)
	}
	if !conn.closed || conn.timeout != ldapTimeout {
		t.Errorf("connection closed = %t, timeout = %v", conn.closed, conn.timeout)
	}
	if strings.Contains(output.String(), testLDAPBindSecret) || strings.Contains(output.String(), testLDAPUserSecret) || strings.Contains(output.String(), "alice@example.com") {
		t.Errorf("sensitive LDAP values appeared in logs: %s", output.String())
	}
}

func TestLDAPLoginWithoutPendingSessionRedirectsToHome(t *testing.T) {
	conn := &fakeLDAPConn{
		userEntry:    ldap.NewEntry(testLDAPUserDN, map[string][]string{"uid": {testLDAPUsername}}),
		groupEntries: []*ldap.Entry{ldap.NewEntry(testLDAPGroupDN, nil)},
	}
	_, _, e := newTestLDAPProvider(t, ldapTestConfig(), conn, &bytes.Buffer{})
	response := postLDAPCredentials(e, nil, testLDAPUsername, testLDAPUserSecret)

	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var result map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["redirect"] != "/" {
		t.Errorf("redirect = %q, want /", result["redirect"])
	}
}

func TestLDAPUserFilterEscapesUsername(t *testing.T) {
	conn := &fakeLDAPConn{
		userEntry:    ldap.NewEntry(testLDAPUserDN, map[string][]string{"uid": {testLDAPUsername}}),
		groupEntries: []*ldap.Entry{ldap.NewEntry(testLDAPGroupDN, nil)},
	}
	provider, _, _ := newTestLDAPProvider(t, ldapTestConfig(), conn, &bytes.Buffer{})

	username := `*)(uid=*)`
	_, failure := provider.authenticate(username, testLDAPUserSecret)
	if failure != "" {
		t.Fatalf("authentication failure = %q", failure)
	}
	want := "(uid=" + ldap.EscapeFilter(username) + ")"
	if got := conn.searchRequests[0].Filter; got != want {
		t.Errorf("user filter = %q, want %q", got, want)
	}
}

func TestLDAPLoginRejectsInvalidCredentialsWithoutLoggingSecrets(t *testing.T) {
	conn := &fakeLDAPConn{
		userEntry:    ldap.NewEntry(testLDAPUserDN, map[string][]string{"uid": {testLDAPUsername}}),
		groupEntries: []*ldap.Entry{ldap.NewEntry(testLDAPGroupDN, nil)},
		userBindErr:  &ldap.Error{ResultCode: ldap.LDAPResultInvalidCredentials, Err: errors.New("password rejected")},
	}
	var output bytes.Buffer
	_, _, e := newTestLDAPProvider(t, ldapTestConfig(), conn, &output)
	response := postLDAPCredentials(e, beginLDAPLogin(t, e), testLDAPUsername, testLDAPUserSecret)

	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "Invalid username or password") {
		t.Fatalf("invalid credential response = %d %q", response.Code, response.Body.String())
	}
	if strings.Contains(output.String(), testLDAPUserSecret) || strings.Contains(output.String(), testLDAPBindSecret) || strings.Contains(output.String(), "password rejected") {
		t.Errorf("LDAP secrets appeared in logs: %s", output.String())
	}
}

func TestLDAPGroupRestrictionAndDirectoryFailures(t *testing.T) {
	tests := []struct {
		name       string
		conn       *fakeLDAPConn
		dialError  error
		wantStatus int
		wantType   string
	}{
		{
			name: "user is not in an allowed group",
			conn: &fakeLDAPConn{
				userEntry:   ldap.NewEntry(testLDAPUserDN, map[string][]string{"uid": {testLDAPUsername}}),
				userBindErr: nil,
			},
			wantStatus: http.StatusForbidden,
			wantType:   "group_restriction_failed",
		},
		{
			name: "group directory search failure",
			conn: &fakeLDAPConn{
				userEntry:      ldap.NewEntry(testLDAPUserDN, map[string][]string{"uid": {testLDAPUsername}}),
				groupSearchErr: errors.New("directory response contains private data"),
			},
			wantStatus: http.StatusServiceUnavailable,
			wantType:   "ldap_group_search_failed",
		},
		{
			name:       "directory connection failure",
			dialError:  errors.New("connection details must not be logged"),
			wantStatus: http.StatusServiceUnavailable,
			wantType:   "ldap_connection_failed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			provider, _, e := newTestLDAPProvider(t, ldapTestConfig(), test.conn, &output)
			if test.dialError != nil {
				provider.dialLDAP = func(string, ...ldap.DialOpt) (ldapConnection, error) {
					return nil, test.dialError
				}
			}
			cookies := beginLDAPLogin(t, e)
			response := postLDAPCredentials(e, cookies, testLDAPUsername, testLDAPUserSecret)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if !strings.Contains(output.String(), `"error.type":"`+test.wantType+`"`) {
				t.Errorf("missing sanitized failure type %q: %s", test.wantType, output.String())
			}
			if strings.Contains(output.String(), testLDAPBindSecret) || strings.Contains(output.String(), testLDAPUserSecret) || strings.Contains(output.String(), "private data") || strings.Contains(output.String(), "connection details") {
				t.Errorf("directory details appeared in logs: %s", output.String())
			}
		})
	}
}

func TestLDAPStartTLSUsesVerifiedServerConfig(t *testing.T) {
	conn := &fakeLDAPConn{}
	ldapConfig := ldapTestConfig()
	var output bytes.Buffer
	provider, _, _ := newTestLDAPProvider(t, ldapConfig, conn, &output)
	if _, err := provider.connect(); err != nil {
		t.Fatal(err)
	}
	if conn.startTLSConfig == nil || conn.startTLSConfig.ServerName != "ldap.example.com" || conn.startTLSConfig.MinVersion != tls.VersionTLS12 || conn.startTLSConfig.InsecureSkipVerify {
		t.Errorf("unexpected StartTLS configuration: %#v", conn.startTLSConfig)
	}
}

func TestLDAPTLSCertificateValidation(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	defer server.Close()
	ldapURL := "ldaps://" + strings.TrimPrefix(server.URL, "https://")

	ldapConfig := ldapTestConfig()
	ldapConfig.URL = ldapURL
	ldapConfig.StartTLS = false
	provider, err := NewLDAPProvider(ldapConfig, newTestSessionManager(), logging.New(logging.Config{Output: &bytes.Buffer{}}))
	if err != nil {
		t.Fatal(err)
	}
	if conn, err := provider.connect(); err == nil {
		conn.Close()
		t.Fatal("self-signed certificate was accepted without a custom CA")
	}

	ldapConfig.InsecureSkipVerify = true
	provider, err = NewLDAPProvider(ldapConfig, newTestSessionManager(), logging.New(logging.Config{Output: &bytes.Buffer{}}))
	if err != nil {
		t.Fatal(err)
	}
	conn, err := provider.connect()
	if err != nil {
		t.Fatalf("explicit certificate verification bypass failed: %v", err)
	}
	conn.Close()
	ldapConfig.InsecureSkipVerify = false

	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: server.Certificate().Raw,
	})
	certificateFile := t.TempDir() + "/ldap-ca.pem"
	if err := os.WriteFile(certificateFile, certificatePEM, 0600); err != nil {
		t.Fatal(err)
	}
	ldapConfig.CACertFile = certificateFile
	provider, err = NewLDAPProvider(ldapConfig, newTestSessionManager(), logging.New(logging.Config{Output: &bytes.Buffer{}}))
	if err != nil {
		t.Fatal(err)
	}
	conn, err = provider.connect()
	if err != nil {
		t.Fatalf("trusted certificate was rejected: %v", err)
	}
	conn.Close()
}
