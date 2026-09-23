package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
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

func NewManager(ctx context.Context, authConfig config.Auth, appURI string, logger *logging.Logger) (*Manager, error) {
	sessions := NewSessionManager(authConfig)
	var provider Provider
	var err error
	switch authConfig.Provider {
	case "azure":
		provider = NewAzureProvider(authConfig, appURI, sessions, logger)
	case "okta":
		provider = NewOktaProvider(authConfig, appURI, sessions, logger)
	case "oidc":
		provider, err = NewOIDCProvider(ctx, authConfig.OIDC, appURI, sessions, logger)
	default:
		err = fmt.Errorf("unsupported authentication provider %q", authConfig.Provider)
	}
	if err != nil {
		return nil, err
	}
	return &Manager{
		provider: provider,
		sessions: sessions,
		appURI:   appURI,
		publicPath: map[string]bool{
			"/health":               true,
			"/api/slack":            true,
			"/logout":               true,
			provider.CallbackPath(): true,
		},
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
	return m.publicPath[path]
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
