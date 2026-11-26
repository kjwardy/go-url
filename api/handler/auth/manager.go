package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-pg/pg"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/kjwardy/go-url/api/utils"
	"github.com/labstack/echo"
)

type Provider interface {
	Name() string
	RegisterRoutes(*echo.Echo)
	BeginLogin(echo.Context) error
	LogoutURL(string) (string, bool)
	CallbackPath() string
}

type Manager struct {
	provider   Provider
	sessions   *SessionManager
	appURI     string
	publicPath map[string]bool
}

func NewManager(ctx context.Context, authConfig config.Auth, appURI string, logger *logging.Logger, database *pg.DB) (*Manager, error) {
	if database == nil {
		return nil, fmt.Errorf("database connection is required for authentication sessions")
	}
	sessions := NewSessionManager(authConfig, database, logger)
	var provider Provider
	var err error
	switch authConfig.Provider {
	case "azure":
		provider = NewAzureProvider(authConfig, appURI, sessions, logger)
	case "okta":
		provider = NewOktaProvider(authConfig, appURI, sessions, logger)
	case "oidc":
		provider, err = NewOIDCProvider(ctx, authConfig.OIDC, appURI, sessions, logger)
	case "ldap":
		provider, err = NewLDAPProvider(authConfig.LDAP, sessions, logger)
	default:
		err = fmt.Errorf("unsupported authentication provider %q", authConfig.Provider)
	}
	if err != nil {
		return nil, err
	}
	publicPath := map[string]bool{
		"/health":               true,
		"/api/slack":            true,
		"/logout":               true,
		provider.CallbackPath(): true,
	}
	if provider.Name() == "ldap" {
		for _, path := range []string{
			"/go/login",
			"/go/config.js",
			"/go/manifest.json",
			"/go/favicon.ico",
			"/go/logo.png",
			"/go/logo.svg",
		} {
			publicPath[path] = true
		}
	}
	return &Manager{
		provider:   provider,
		sessions:   sessions,
		appURI:     appURI,
		publicPath: publicPath,
	}, nil
}

func (m *Manager) RegisterRoutes(e *echo.Echo) {
	m.provider.RegisterRoutes(e)
	e.POST("/logout", m.logout)
}

func (m *Manager) Authenticate(next echo.HandlerFunc, c echo.Context) error {
	if m.sessions.IsAuthenticated(c) {
		return next(c)
	}
	return m.provider.BeginLogin(c)
}

func (m *Manager) IsPublic(path string) bool {
	return m.publicPath[path] || (m.provider.Name() == "ldap" && strings.HasPrefix(path, "/go/static/"))
}

func (m *Manager) Identity(c echo.Context) (*Identity, error) {
	return m.sessions.Identity(c)
}

func requestSourceIP(c echo.Context) string {
	authConfig := config.GetConfig().Auth
	return utils.ClientIP(
		c.Request().RemoteAddr,
		c.Request().Header.Get("X-Forwarded-For"),
		authConfig.AllowForwardedFor,
		authConfig.ForwardedForTrustLevel,
	)
}

func randomValue(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func secureEqual(left, right string) bool {
	if len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func (m *Manager) logout(c echo.Context) error {
	idToken, err := m.sessions.Logout(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to log out")
	}
	if logoutURL, ok := m.provider.LogoutURL(idToken); ok {
		return c.Redirect(http.StatusSeeOther, logoutURL)
	}
	return c.Redirect(http.StatusSeeOther, m.appURI)
}
