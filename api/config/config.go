package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/kelseyhightower/envconfig"
)

// Specification definition in env
type Specification struct {
	Debug                  bool
	JSONLogs               bool     `envconfig:"JSON_LOGS"`
	Port                   int      `default:"1323"`
	EnableAuth             bool     `envconfig:"ENABLE_AUTH"`
	AuthProvider           string   `envconfig:"AUTH_PROVIDER"`
	AuthExpirySeconds      int      `envconfig:"AUTH_EXPIRY_SECONDS" default:"2592000"`
	SecureCookies          bool     `envconfig:"SECURE_COOKIES" default:"true"`
	ADTenantID             string   `envconfig:"AD_TENANT_ID"`
	ADClientID             string   `envconfig:"AD_CLIENT_ID"`
	ADClientSecret         string   `envconfig:"AD_CLIENT_SECRET"`
	SessionToken           string   `envconfig:"SESSION_TOKEN"`
	PostgresAddr           string   `envconfig:"POSTGRES_ADDR" default:"localhost:5432"`
	PostgresDatabase       string   `envconfig:"POSTGRES_DATABASE" default:"go"`
	PostgresUser           string   `envconfig:"POSTGRES_USER" default:"postgres"`
	PostgresPass           string   `envconfig:"POSTGRES_PASS" default:"password"`
	Hosts                  []string `required:"true"`
	BlockedHosts           []string `envconfig:"BLOCKED_HOSTS"`
	AppURI                 string   `envconfig:"APP_URI" required:"true"`
	SlackToken             string   `envconfig:"SLACK_TOKEN"`
	SlackSigningSecret     string   `envconfig:"SLACK_SIGNING_SECRET"`
	SlackTeamID            string   `envconfig:"SLACK_TEAM_ID"`
	AllowedIPs             []string `envconfig:"ALLOWED_IPS"`
	AllowForwardedFor      bool     `envconfig:"ALLOW_FORWARDED_FOR"`
	ForwardedForTrustLevel int      `envconfig:"FORWARDED_FOR_TRUST_LEVEL" default:"1"`
	SentryDSN              string   `envconfig:"SENTRY_API_DSN"`
	ServiceVersion         string   `envconfig:"SERVICE_VERSION"`
	ServiceEnvironment     string   `envconfig:"SERVICE_ENVIRONMENT"`
	OktaClientID           string   `envconfig:"OKTA_CLIENT_ID"`
	OktaClientSecret       string   `envconfig:"OKTA_CLIENT_SECRET"`
	OktaIssuer             string   `envconfig:"OKTA_ISSUER"`
	OIDCIssuerURL          string   `envconfig:"OIDC_ISSUER_URL"`
	OIDCClientID           string   `envconfig:"OIDC_CLIENT_ID"`
	OIDCClientSecret       string   `envconfig:"OIDC_CLIENT_SECRET"`
	OIDCRedirectURL        string   `envconfig:"OIDC_REDIRECT_URL"`
	OIDCScopes             []string `envconfig:"OIDC_SCOPES"`
	OIDCUsernameClaim      string   `envconfig:"OIDC_USERNAME_CLAIM"`
	OIDCEmailClaim         string   `envconfig:"OIDC_EMAIL_CLAIM"`
	LDAPURL                string   `envconfig:"LDAP_URL"`
	LDAPStartTLS           bool     `envconfig:"LDAP_STARTTLS"`
	LDAPInsecureSkipVerify bool     `envconfig:"LDAP_INSECURE_SKIP_VERIFY"`
	LDAPCACertFile         string   `envconfig:"LDAP_CA_CERT_FILE"`
	LDAPBindDN             string   `envconfig:"LDAP_BIND_DN"`
	LDAPBindPassword       string   `envconfig:"LDAP_BIND_PASSWORD"`
	LDAPUserSearchBase     string   `envconfig:"LDAP_USER_SEARCH_BASE"`
	LDAPUserFilter         string   `envconfig:"LDAP_USER_FILTER"`
	LDAPUsernameAttribute  string   `envconfig:"LDAP_USERNAME_ATTRIBUTE"`
	LDAPGroupSearchBase    string   `envconfig:"LDAP_GROUP_SEARCH_BASE"`
	LDAPGroupFilter        string   `envconfig:"LDAP_GROUP_FILTER"`
	LDAPAllowedGroups      string   `envconfig:"LDAP_ALLOWED_GROUPS"`
}

type OIDC struct {
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	UsernameClaim string
	EmailClaim    string
}

type LDAP struct {
	URL                string
	StartTLS           bool
	InsecureSkipVerify bool
	CACertFile         string
	BindDN             string
	BindPassword       string
	UserSearchBase     string
	UserFilter         string
	UsernameAttribute  string
	GroupSearchBase    string
	GroupFilter        string
	AllowedGroups      []string
}

// Auth config
type Auth struct {
	Enabled                bool
	Provider               string
	ADTenantID             string
	ADClientID             string
	ADClientSecret         string
	SessionToken           string
	AllowedIPs             []string
	AllowForwardedFor      bool
	ForwardedForTrustLevel int
	MaxAge                 int
	SecureCookies          bool
	OktaClientID           string
	OktaClientSecret       string
	OktaIssuer             string
	OIDC                   OIDC
	LDAP                   LDAP
}

// Database config
type Database struct {
	Addr     string
	Database string
	User     string
	Pass     string
}

// Slack config
type Slack struct {
	Token         string
	SigningSecret string
	TeamID        string
}

// Config definition
type Config struct {
	Debug              bool
	JSONLogs           bool
	Port               int
	Auth               Auth
	Database           Database
	AppURI             string
	BlockedHosts       []string
	Slack              Slack
	SentryDSN          string
	ServiceVersion     string
	ServiceEnvironment string
}

func Validate(c Config) error {
	if c.Auth.ForwardedForTrustLevel < 1 {
		return fmt.Errorf("FORWARDED_FOR_TRUST_LEVEL must be greater than 0")
	}
	provider := strings.ToLower(strings.TrimSpace(c.Auth.Provider))
	configured := map[string]bool{
		"azure": c.Auth.ADTenantID != "" || c.Auth.ADClientID != "" || c.Auth.ADClientSecret != "",
		"okta":  c.Auth.OktaClientID != "" || c.Auth.OktaClientSecret != "" || c.Auth.OktaIssuer != "",
		"oidc":  oidcConfigured(c.Auth.OIDC),
		"ldap":  ldapConfigured(c.Auth.LDAP),
	}
	if !c.Auth.Enabled {
		if provider != "" {
			return fmt.Errorf("AUTH_PROVIDER must be empty when authentication is disabled")
		}
		for name, active := range configured {
			if active {
				return fmt.Errorf("%s authentication configuration is set while authentication is disabled", name)
			}
		}
		return nil
	}
	if c.Auth.SessionToken == "" {
		return fmt.Errorf("SESSION_TOKEN is required when authentication is enabled")
	}
	if c.Auth.MaxAge < 1 {
		return fmt.Errorf("AUTH_EXPIRY_SECONDS must be greater than 0 when authentication is enabled")
	}
	if provider != "azure" && provider != "okta" && provider != "oidc" && provider != "ldap" {
		return fmt.Errorf("AUTH_PROVIDER must be azure, okta, oidc, or ldap when authentication is enabled")
	}
	for name, active := range configured {
		if name != provider && active {
			return fmt.Errorf("%s authentication configuration conflicts with AUTH_PROVIDER=%s", name, provider)
		}
	}
	switch provider {
	case "azure":
		if c.Auth.ADTenantID == "" || c.Auth.ADClientID == "" || c.Auth.ADClientSecret == "" {
			return fmt.Errorf("AD_TENANT_ID, AD_CLIENT_ID, and AD_CLIENT_SECRET are required for Azure authentication")
		}
	case "okta":
		if c.Auth.OktaClientID == "" || c.Auth.OktaClientSecret == "" || c.Auth.OktaIssuer == "" {
			return fmt.Errorf("OKTA_CLIENT_ID, OKTA_CLIENT_SECRET, and OKTA_ISSUER are required for Okta authentication")
		}
	case "oidc":
		if err := validateOIDC(c.Auth.OIDC); err != nil {
			return err
		}
	case "ldap":
		if err := validateLDAP(c.Auth.LDAP, c.AppURI); err != nil {
			return err
		}
	}
	return nil
}

func oidcConfigured(c OIDC) bool {
	return c.IssuerURL != "" || c.ClientID != "" || c.ClientSecret != "" || c.RedirectURL != "" || len(c.Scopes) > 0 || c.UsernameClaim != "" || c.EmailClaim != ""
}

func ldapConfigured(c LDAP) bool {
	return c.URL != "" ||
		c.StartTLS ||
		c.InsecureSkipVerify ||
		c.CACertFile != "" ||
		c.BindDN != "" ||
		c.BindPassword != "" ||
		c.UserSearchBase != "" ||
		c.UserFilter != "" ||
		c.UsernameAttribute != "" ||
		c.GroupSearchBase != "" ||
		c.GroupFilter != "" ||
		len(c.AllowedGroups) > 0
}

func validateLDAP(c LDAP, appURI string) error {
	if c.URL == "" || c.BindDN == "" || c.BindPassword == "" || c.UserSearchBase == "" {
		return fmt.Errorf("LDAP_URL, LDAP_BIND_DN, LDAP_BIND_PASSWORD, and LDAP_USER_SEARCH_BASE are required")
	}
	serverURL, err := url.Parse(c.URL)
	if err != nil || serverURL.Host == "" || serverURL.User != nil || serverURL.RawQuery != "" || serverURL.Fragment != "" || (serverURL.Path != "" && serverURL.Path != "/") {
		return fmt.Errorf("LDAP_URL must be an LDAP server URL without credentials or a path")
	}
	if serverURL.Scheme != "ldap" && serverURL.Scheme != "ldaps" {
		return fmt.Errorf("LDAP_URL must use ldap:// or ldaps://")
	}
	if serverURL.Scheme == "ldaps" && c.StartTLS {
		return fmt.Errorf("LDAP_STARTTLS cannot be enabled with an ldaps:// LDAP_URL")
	}
	if err := validateAuthURL("APP_URI", appURI, false); err != nil {
		return err
	}
	if !validLDAPAttribute(c.UsernameAttribute) {
		return fmt.Errorf("LDAP_USERNAME_ATTRIBUTE must be a valid LDAP attribute name")
	}
	if strings.Count(c.UserFilter, "{username}") != 1 {
		return fmt.Errorf("LDAP_USER_FILTER must contain exactly one {username} placeholder")
	}
	if _, err := ldap.CompileFilter(strings.Replace(c.UserFilter, "{username}", "go-url-user", 1)); err != nil {
		return fmt.Errorf("LDAP_USER_FILTER is invalid")
	}
	if len(c.AllowedGroups) == 0 {
		if c.GroupSearchBase != "" || c.GroupFilter != "" {
			return fmt.Errorf("LDAP_GROUP_SEARCH_BASE and LDAP_GROUP_FILTER require LDAP_ALLOWED_GROUPS")
		}
		return nil
	}
	if c.GroupSearchBase == "" || strings.Count(c.GroupFilter, "{user_dn}") != 1 {
		return fmt.Errorf("LDAP_GROUP_SEARCH_BASE and LDAP_GROUP_FILTER with one {user_dn} placeholder are required when LDAP_ALLOWED_GROUPS is set")
	}
	if _, err := ldap.CompileFilter(strings.Replace(c.GroupFilter, "{user_dn}", "cn=go-url-user,dc=example,dc=com", 1)); err != nil {
		return fmt.Errorf("LDAP_GROUP_FILTER is invalid")
	}
	for _, groupDN := range c.AllowedGroups {
		if _, err := ldap.ParseDN(groupDN); err != nil {
			return fmt.Errorf("LDAP_ALLOWED_GROUPS contains an invalid distinguished name")
		}
	}
	return nil
}

func validLDAPAttribute(attribute string) bool {
	if attribute == "" || !((attribute[0] >= 'a' && attribute[0] <= 'z') || (attribute[0] >= 'A' && attribute[0] <= 'Z')) {
		return false
	}
	for _, character := range attribute[1:] {
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			character == '-') {
			return false
		}
	}
	return true
}

func validateOIDC(c OIDC) error {
	if c.IssuerURL == "" || c.ClientID == "" || c.ClientSecret == "" || c.RedirectURL == "" {
		return fmt.Errorf("OIDC_ISSUER_URL, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and OIDC_REDIRECT_URL are required for OIDC authentication")
	}
	if err := validateAuthURL("OIDC_ISSUER_URL", c.IssuerURL, false); err != nil {
		return err
	}
	if err := validateAuthURL("OIDC_REDIRECT_URL", c.RedirectURL, true); err != nil {
		return err
	}
	if !contains(c.Scopes, "openid") {
		return fmt.Errorf("OIDC_SCOPES must include openid")
	}
	if c.UsernameClaim == "" {
		return fmt.Errorf("OIDC_USERNAME_CLAIM must not be empty")
	}
	return nil
}

func validateAuthURL(name, value string, callback bool) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("%s must be an absolute URL", name)
	}
	if parsed.Scheme != "https" {
		host := parsed.Hostname()
		ip := net.ParseIP(host)
		if parsed.Scheme != "http" || (host != "localhost" && (ip == nil || !ip.IsLoopback())) {
			return fmt.Errorf("%s must use HTTPS except for loopback development URLs", name)
		}
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%s must not include a query string or fragment", name)
	}
	if callback && parsed.Path != "/oidc/callback" {
		return fmt.Errorf("OIDC_REDIRECT_URL path must be /oidc/callback")
	}
	return nil
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

var config = Config{}

// Init loads the config from the env and sets defaults
func Init() error {
	var spec Specification

	err := envconfig.Process("", &spec)
	if err != nil {
		return err
	}
	config.Debug = spec.Debug
	config.Port = spec.Port
	config.JSONLogs = spec.JSONLogs
	oidcScopes := spec.OIDCScopes
	if len(oidcScopes) == 0 && strings.EqualFold(spec.AuthProvider, "oidc") {
		oidcScopes = []string{"openid", "profile", "email"}
	}
	oidcUsernameClaim := spec.OIDCUsernameClaim
	if oidcUsernameClaim == "" && strings.EqualFold(spec.AuthProvider, "oidc") {
		oidcUsernameClaim = "preferred_username"
	}
	oidcEmailClaim := spec.OIDCEmailClaim
	if oidcEmailClaim == "" && strings.EqualFold(spec.AuthProvider, "oidc") {
		oidcEmailClaim = "email"
	}
	ldapUsernameAttribute := spec.LDAPUsernameAttribute
	ldapUserFilter := spec.LDAPUserFilter
	if strings.EqualFold(spec.AuthProvider, "ldap") {
		if ldapUsernameAttribute == "" {
			ldapUsernameAttribute = "uid"
		}
		if ldapUserFilter == "" {
			ldapUserFilter = "(" + ldapUsernameAttribute + "={username})"
		}
	}
	var ldapAllowedGroups []string
	if spec.LDAPAllowedGroups != "" {
		for _, groupDN := range strings.Split(spec.LDAPAllowedGroups, ";") {
			ldapAllowedGroups = append(ldapAllowedGroups, strings.TrimSpace(groupDN))
		}
	}
	config.Auth = Auth{
		Enabled:                spec.EnableAuth,
		Provider:               strings.ToLower(strings.TrimSpace(spec.AuthProvider)),
		ADTenantID:             spec.ADTenantID,
		ADClientID:             spec.ADClientID,
		ADClientSecret:         spec.ADClientSecret,
		SessionToken:           spec.SessionToken,
		AllowedIPs:             spec.AllowedIPs,
		AllowForwardedFor:      spec.AllowForwardedFor,
		ForwardedForTrustLevel: spec.ForwardedForTrustLevel,
		MaxAge:                 spec.AuthExpirySeconds,
		SecureCookies:          spec.SecureCookies,
		OktaClientID:           spec.OktaClientID,
		OktaClientSecret:       spec.OktaClientSecret,
		OktaIssuer:             spec.OktaIssuer,
		OIDC: OIDC{
			IssuerURL:     spec.OIDCIssuerURL,
			ClientID:      spec.OIDCClientID,
			ClientSecret:  spec.OIDCClientSecret,
			RedirectURL:   spec.OIDCRedirectURL,
			Scopes:        oidcScopes,
			UsernameClaim: oidcUsernameClaim,
			EmailClaim:    oidcEmailClaim,
		},
		LDAP: LDAP{
			URL:                spec.LDAPURL,
			StartTLS:           spec.LDAPStartTLS,
			InsecureSkipVerify: spec.LDAPInsecureSkipVerify,
			CACertFile:         spec.LDAPCACertFile,
			BindDN:             spec.LDAPBindDN,
			BindPassword:       spec.LDAPBindPassword,
			UserSearchBase:     spec.LDAPUserSearchBase,
			UserFilter:         ldapUserFilter,
			UsernameAttribute:  ldapUsernameAttribute,
			GroupSearchBase:    spec.LDAPGroupSearchBase,
			GroupFilter:        spec.LDAPGroupFilter,
			AllowedGroups:      ldapAllowedGroups,
		},
	}
	config.Database = Database{
		Addr:     spec.PostgresAddr,
		Database: spec.PostgresDatabase,
		User:     spec.PostgresUser,
		Pass:     spec.PostgresPass,
	}
	config.AppURI = spec.AppURI

	config.Slack = Slack{
		Token:         spec.SlackToken,
		SigningSecret: spec.SlackSigningSecret,
		TeamID:        spec.SlackTeamID,
	}

	config.SentryDSN = spec.SentryDSN
	config.ServiceVersion = spec.ServiceVersion
	config.ServiceEnvironment = spec.ServiceEnvironment

	config.BlockedHosts = append(spec.BlockedHosts, spec.Hosts...)

	if err := Validate(config); err != nil {
		return err
	}
	return nil
}

// GetConfig returns the config
func GetConfig() Config {
	return config
}
