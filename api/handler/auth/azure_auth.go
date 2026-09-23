package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

type AzureProvider struct {
	oauthConfig *oauth2.Config
	sessions    *SessionManager
	logger      *logging.Logger
}

type AzureUser struct {
	ID                string `json:"id"`
	Email             string `json:"mail"`
	DisplayName       string `json:"displayName"`
	UserPrincipalName string `json:"userPrincipalName"`
}

func NewAzureProvider(authConfig config.Auth, appURI string, sessions *SessionManager, logger *logging.Logger) *AzureProvider {
	endpointURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0", authConfig.ADTenantID)
	return &AzureProvider{
		oauthConfig: &oauth2.Config{
			ClientID:     authConfig.ADClientID,
			ClientSecret: authConfig.ADClientSecret,
			RedirectURL:  appURI + "/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  endpointURL + "/authorize",
				TokenURL: endpointURL + "/token",
			},
			Scopes: []string{"User.Read"},
		},
		sessions: sessions,
		logger:   logger,
	}
}

func (p *AzureProvider) Name() string {
	return "azure"
}

func (p *AzureProvider) CallbackPath() string {
	return "/callback"
}

func (p *AzureProvider) RegisterRoutes(e *echo.Echo) {
	e.GET(p.CallbackPath(), p.callback)
}

func (p *AzureProvider) BeginLogin(c echo.Context) error {
	state, err := randomValue(32)
	if err != nil {
		p.failure(c, "state_generation_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start authentication")
	}
	if err := p.sessions.BeginLogin(c, &PendingLogin{Provider: p.Name(), State: state, RedirectPath: c.Request().URL.RequestURI()}); err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start authentication")
	}
	return c.Redirect(http.StatusTemporaryRedirect, p.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline))
}

func (p *AzureProvider) callback(c echo.Context) error {
	pending, err := p.sessions.ConsumePendingLogin(c, p.Name())
	if err != nil || !secureEqual(c.QueryParam("state"), pending.State) {
		p.failure(c, "invalid_state")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid callback state")
	}
	code := c.QueryParam("code")
	if code == "" {
		p.failure(c, "missing_code")
		return echo.NewHTTPError(http.StatusBadRequest, "Authorization code is missing")
	}
	token, err := p.oauthConfig.Exchange(c.Request().Context(), code)
	if err != nil {
		p.failure(c, "token_exchange_failed")
		return echo.NewHTTPError(http.StatusBadGateway, "Authentication provider rejected the callback")
	}
	user, err := getAzureUser(c.Request().Context(), p.oauthConfig.Client(c.Request().Context(), token))
	if err != nil {
		p.failure(c, "user_details_failed")
		return echo.NewHTTPError(http.StatusBadGateway, "Unable to retrieve user identity")
	}
	username := user.Email
	if username == "" {
		username = user.UserPrincipalName
	}
	if username == "" {
		p.failure(c, "missing_username_claim")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authenticated identity has no username")
	}
	redirect, err := p.sessions.CompleteLogin(c, &Identity{Provider: p.Name(), Subject: user.ID, Username: username, Email: user.Email}, "", pending.RedirectPath)
	if err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create authenticated session")
	}
	return c.Redirect(http.StatusTemporaryRedirect, redirect)
}

func (p *AzureProvider) LogoutURL(string) (string, bool) {
	return "", false
}

func (p *AzureProvider) failure(c echo.Context, errorType string) {
	p.logger.AuthenticationFailure(logging.RequestID(c), p.Name(), errorType)
}

func getAzureUser(ctx context.Context, client *http.Client) (*AzureUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user details response status %d", resp.StatusCode)
	}
	var user AzureUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
