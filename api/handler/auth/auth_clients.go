package auth

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/kjwardy/go-url/api/config"
	"github.com/labstack/echo"
)

const sessionName = "session"

type Identity struct {
	Provider string
	Subject  string
	Username string
	Email    string
}

type PendingLogin struct {
	Provider     string
	State        string
	Nonce        string
	PKCEVerifier string
	RedirectPath string
}

type SessionManager struct {
	store         sessions.Store
	maxAge        int
	secureCookies bool
}

var registerSessionTypes sync.Once

func NewSessionManager(authConfig config.Auth) *SessionManager {
	store := sessions.NewFilesystemStore("", []byte(authConfig.SessionToken), nil)
	store.MaxLength(0)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   authConfig.MaxAge,
		HttpOnly: true,
		Secure:   authConfig.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	}
	store.MaxAge(authConfig.MaxAge)
	return newSessionManager(store, authConfig.MaxAge, authConfig.SecureCookies)
}

func newSessionManager(store sessions.Store, maxAge int, secureCookies bool) *SessionManager {
	registerSessionTypes.Do(func() {
		gob.Register(&Identity{})
		gob.Register(&PendingLogin{})
	})
	return &SessionManager{store: store, maxAge: maxAge, secureCookies: secureCookies}
}

func (m *SessionManager) session(c echo.Context) (*sessions.Session, error) {
	return m.store.Get(c.Request(), sessionName)
}

func (m *SessionManager) IsAuthenticated(c echo.Context) bool {
	session, err := m.session(c)
	return err == nil && session.Values["authenticated"] == true && session.Values["identity"] != nil
}

func (m *SessionManager) Identity(c echo.Context) (*Identity, error) {
	session, err := m.session(c)
	if err != nil {
		return nil, err
	}
	identity, ok := session.Values["identity"].(*Identity)
	if !ok || identity == nil || session.Values["authenticated"] != true {
		return nil, fmt.Errorf("session is not authenticated")
	}
	return identity, nil
}

func (m *SessionManager) BeginLogin(c echo.Context, pending *PendingLogin) error {
	pending.RedirectPath = safeRedirectPath(pending.RedirectPath)
	session, err := m.session(c)
	if err != nil {
		return err
	}
	session.Values["pending_login"] = pending
	return session.Save(c.Request(), c.Response())
}

func (m *SessionManager) ConsumePendingLogin(c echo.Context, provider string) (*PendingLogin, error) {
	session, err := m.session(c)
	if err != nil {
		return nil, err
	}
	pending, ok := session.Values["pending_login"].(*PendingLogin)
	if !ok || pending == nil || pending.Provider != provider {
		return nil, fmt.Errorf("login session is missing or invalid")
	}
	delete(session.Values, "pending_login")
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return nil, err
	}
	return pending, nil
}

func (m *SessionManager) CompleteLogin(c echo.Context, identity *Identity, idToken, redirectPath string) (string, error) {
	oldSession, err := m.session(c)
	if err != nil {
		return "", err
	}
	redirectPath = safeRedirectPath(redirectPath)
	oldSession.Options.MaxAge = -1
	oldSession.Values = map[interface{}]interface{}{}
	if err := oldSession.Save(c.Request(), c.Response()); err != nil {
		return "", err
	}

	session := sessions.NewSession(m.store, sessionName)
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   m.maxAge,
		HttpOnly: true,
		Secure:   m.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
	session.IsNew = true
	session.Values["authenticated"] = true
	session.Values["identity"] = identity
	if idToken != "" {
		session.Values["id_token"] = idToken
	}
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return "", err
	}
	m.setUserCookie(c, identity.Username, m.maxAge)
	return redirectPath, nil
}

func (m *SessionManager) Logout(c echo.Context) (string, error) {
	session, err := m.session(c)
	if err != nil {
		return "", err
	}
	idToken, _ := session.Values["id_token"].(string)
	session.Options.MaxAge = -1
	session.Values = map[interface{}]interface{}{}
	if err := session.Save(c.Request(), c.Response()); err != nil {
		return "", err
	}
	m.setUserCookie(c, "", -1)
	return idToken, nil
}

func (m *SessionManager) setUserCookie(c echo.Context, username string, maxAge int) {
	cookie := &http.Cookie{
		Name:     "user",
		Value:    url.PathEscape(username),
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: false,
		Secure:   m.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}
	if maxAge < 0 {
		cookie.Expires = time.Unix(1, 0)
	}
	c.SetCookie(cookie)
}

func safeRedirectPath(path string) string {
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "/"
	}
	parsed, err := url.Parse(path)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return "/"
	}
	return parsed.RequestURI()
}
