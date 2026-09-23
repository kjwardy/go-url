package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	oauthConfig        *oauth2.Config
	verifier           *oidc.IDTokenVerifier
	sessions           *SessionManager
	logger             *logging.Logger
	usernameClaim      string
	emailClaim         string
	endSessionEndpoint string
	appURI             string
}

func NewOIDCProvider(ctx context.Context, oidcConfig config.OIDC, appURI string, sessions *SessionManager, logger *logging.Logger) (*OIDCProvider, error) {
	discoveryContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	provider, err := oidc.NewProvider(discoveryContext, oidcConfig.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery failed: %w", err)
	}
	var metadata struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := provider.Claims(&metadata); err != nil {
		return nil, fmt.Errorf("OIDC discovery metadata was invalid: %w", err)
	}
	return &OIDCProvider{
		oauthConfig: &oauth2.Config{
			ClientID:     oidcConfig.ClientID,
			ClientSecret: oidcConfig.ClientSecret,
			RedirectURL:  oidcConfig.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       oidcConfig.Scopes,
		},
		verifier:           provider.Verifier(&oidc.Config{ClientID: oidcConfig.ClientID}),
		sessions:           sessions,
		logger:             logger,
		usernameClaim:      oidcConfig.UsernameClaim,
		emailClaim:         oidcConfig.EmailClaim,
		endSessionEndpoint: metadata.EndSessionEndpoint,
		appURI:             appURI,
	}, nil
}

func (p *OIDCProvider) Name() string {
	return "oidc"
}

func (p *OIDCProvider) CallbackPath() string {
	return "/oidc/callback"
}

func (p *OIDCProvider) RegisterRoutes(e *echo.Echo) {
	e.GET(p.CallbackPath(), p.callback)
}

func (p *OIDCProvider) BeginLogin(c echo.Context) error {
	state, err := randomValue(32)
	if err != nil {
		p.failure(c, "state_generation_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start authentication")
	}
	nonce, err := randomValue(32)
	if err != nil {
		p.failure(c, "nonce_generation_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start authentication")
	}
	pkceVerifier := oauth2.GenerateVerifier()
	pending := &PendingLogin{
		Provider:     p.Name(),
		State:        state,
		Nonce:        nonce,
		PKCEVerifier: pkceVerifier,
		RedirectPath: c.Request().URL.RequestURI(),
	}
	if err := p.sessions.BeginLogin(c, pending); err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start authentication")
	}
	authorizationURL := p.oauthConfig.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.S256ChallengeOption(pkceVerifier),
	)
	return c.Redirect(http.StatusTemporaryRedirect, authorizationURL)
}

func (p *OIDCProvider) callback(c echo.Context) error {
	if c.QueryParam("error") != "" {
		_, _ = p.sessions.ConsumePendingLogin(c, p.Name())
		p.failure(c, "provider_error")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication provider rejected the request")
	}
	pending, err := p.sessions.ConsumePendingLogin(c, p.Name())
	if err != nil {
		p.failure(c, "missing_login_session")
		return echo.NewHTTPError(http.StatusBadRequest, "Login session is missing or expired")
	}
	if !secureEqual(c.QueryParam("state"), pending.State) {
		p.failure(c, "invalid_state")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid callback state")
	}
	code := c.QueryParam("code")
	if code == "" {
		p.failure(c, "missing_code")
		return echo.NewHTTPError(http.StatusBadRequest, "Authorization code is missing")
	}
	token, err := p.oauthConfig.Exchange(c.Request().Context(), code, oauth2.VerifierOption(pending.PKCEVerifier))
	if err != nil {
		p.failure(c, "token_exchange_failed")
		return echo.NewHTTPError(http.StatusBadGateway, "Authentication provider rejected the callback")
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		p.failure(c, "missing_id_token")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication provider returned no ID token")
	}
	idToken, err := p.verifier.Verify(c.Request().Context(), rawIDToken)
	if err != nil {
		p.failure(c, "token_validation_failed")
		return echo.NewHTTPError(http.StatusUnauthorized, "ID token validation failed")
	}
	if idToken.Subject == "" {
		p.failure(c, "missing_subject")
		return echo.NewHTTPError(http.StatusUnauthorized, "ID token subject is missing")
	}
	if !secureEqual(idToken.Nonce, pending.Nonce) {
		p.failure(c, "invalid_nonce")
		return echo.NewHTTPError(http.StatusUnauthorized, "ID token nonce validation failed")
	}
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		p.failure(c, "invalid_claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "ID token claims are invalid")
	}
	username, ok := stringClaim(claims, p.usernameClaim)
	if !ok {
		p.failure(c, "missing_username_claim")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authenticated identity has no username")
	}
	email, _ := stringClaim(claims, p.emailClaim)
	redirect, err := p.sessions.CompleteLogin(c, &Identity{
		Provider: p.Name(),
		Subject:  idToken.Subject,
		Username: username,
		Email:    email,
	}, rawIDToken, pending.RedirectPath)
	if err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create authenticated session")
	}
	return c.Redirect(http.StatusTemporaryRedirect, redirect)
}

func (p *OIDCProvider) LogoutURL(idToken string) (string, bool) {
	if p.endSessionEndpoint == "" {
		return "", false
	}
	logoutURL, err := url.Parse(p.endSessionEndpoint)
	if err != nil || !validProviderURL(logoutURL) {
		return "", false
	}
	query := logoutURL.Query()
	query.Set("post_logout_redirect_uri", p.appURI)
	if idToken != "" {
		query.Set("id_token_hint", idToken)
	}
	logoutURL.RawQuery = query.Encode()
	return logoutURL.String(), true
}

func validProviderURL(value *url.URL) bool {
	if !value.IsAbs() || value.Host == "" || value.User != nil {
		return false
	}
	if value.Scheme == "https" {
		return true
	}
	host := value.Hostname()
	ip := net.ParseIP(host)
	return value.Scheme == "http" && (host == "localhost" || (ip != nil && ip.IsLoopback()))
}

func (p *OIDCProvider) failure(c echo.Context, errorType string) {
	p.logger.AuthenticationFailure(logging.RequestID(c), p.Name(), errorType)
}

func stringClaim(claims map[string]interface{}, name string) (string, bool) {
	value, ok := claims[name].(string)
	return value, ok && value != ""
}
