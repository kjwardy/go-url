package auth

import (
	"fmt"
	"net/http"

	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
	verifier "github.com/okta/okta-jwt-verifier-golang"
	"golang.org/x/oauth2"
)

type OktaProvider struct {
	clientID    string
	issuer      string
	oauthConfig *oauth2.Config
	sessions    *SessionManager
	logger      *logging.Logger
}

func NewOktaProvider(authConfig config.Auth, appURI string, sessions *SessionManager, logger *logging.Logger) *OktaProvider {
	return &OktaProvider{
		clientID: authConfig.OktaClientID,
		issuer:   authConfig.OktaIssuer,
		oauthConfig: &oauth2.Config{
			ClientID:     authConfig.OktaClientID,
			ClientSecret: authConfig.OktaClientSecret,
			RedirectURL:  appURI + "/okta/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL:  authConfig.OktaIssuer + "/v1/authorize",
				TokenURL: authConfig.OktaIssuer + "/v1/token",
			},
			Scopes: []string{"openid", "profile", "email"},
		},
		sessions: sessions,
		logger:   logger,
	}
}

func (p *OktaProvider) Name() string {
	return "okta"
}

func (p *OktaProvider) CallbackPath() string {
	return "/okta/callback"
}

func (p *OktaProvider) RegisterRoutes(e *echo.Echo) {
	e.GET(p.CallbackPath(), p.callback)
}

func (p *OktaProvider) BeginLogin(c echo.Context) error {
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
	if err := p.sessions.BeginLogin(c, &PendingLogin{Provider: p.Name(), State: state, Nonce: nonce, RedirectPath: c.Request().URL.RequestURI()}); err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start authentication")
	}
	return c.Redirect(http.StatusTemporaryRedirect, p.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce)))
}

func (p *OktaProvider) callback(c echo.Context) error {
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
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		p.failure(c, "missing_id_token")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication provider returned no ID token")
	}
	verified, err := p.verifyToken(rawIDToken, pending.Nonce)
	if err != nil {
		p.failure(c, "token_validation_failed")
		return echo.NewHTTPError(http.StatusUnauthorized, "ID token validation failed")
	}
	subject, _ := verified.Claims["sub"].(string)
	email, _ := verified.Claims["email"].(string)
	username, _ := verified.Claims["preferred_username"].(string)
	if username == "" {
		username = email
	}
	if username == "" {
		p.failure(c, "missing_username_claim")
		return echo.NewHTTPError(http.StatusUnauthorized, "Authenticated identity has no username")
	}
	redirect, err := p.sessions.CompleteLogin(c, &Identity{Provider: p.Name(), Subject: subject, Username: username, Email: email}, rawIDToken, pending.RedirectPath)
	if err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create authenticated session")
	}
	return c.Redirect(http.StatusTemporaryRedirect, redirect)
}

func (p *OktaProvider) verifyToken(token, nonce string) (*verifier.Jwt, error) {
	jwtVerifier := verifier.JwtVerifier{
		Issuer: p.issuer,
		ClaimsToValidate: map[string]string{
			"nonce": nonce,
			"aud":   p.clientID,
		},
	}
	result, err := jwtVerifier.New().VerifyIdToken(token)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("token could not be verified")
	}
	return result, nil
}

func (p *OktaProvider) LogoutURL(string) (string, bool) {
	return "", false
}

func (p *OktaProvider) failure(c echo.Context, errorType string) {
	p.logger.AuthenticationFailure(logging.RequestID(c), p.Name(), errorType)
}
