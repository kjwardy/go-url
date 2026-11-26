package auth

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

const (
	ldapLoginPath       = "/api/ldap/login"
	ldapLoginPage       = "/go/login"
	maxLDAPLoginBody    = 8192
	maxLDAPUsernameSize = 512
	maxLDAPPasswordSize = 4096
	ldapTimeout         = 5 * time.Second
)

type ldapConnection interface {
	Bind(username, password string) error
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	StartTLS(config *tls.Config) error
	SetTimeout(timeout time.Duration)
	Close() error
}

type ldapDialFunc func(string, ...ldap.DialOpt) (ldapConnection, error)

type LDAPProvider struct {
	serverURL         string
	startTLS          bool
	tlsConfig         *tls.Config
	bindDN            string
	bindPassword      string
	userSearchBase    string
	userFilter        string
	usernameAttribute string
	groupSearchBase   string
	groupFilter       string
	allowedGroups     []*ldap.DN
	sessions          *SessionManager
	logger            *logging.Logger
	dialLDAP          ldapDialFunc
}

type ldapLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewLDAPProvider(ldapConfig config.LDAP, sessions *SessionManager, logger *logging.Logger) (*LDAPProvider, error) {
	tlsConfig, err := newLDAPTLSConfig(ldapConfig)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(ldapConfig.URL, "ldap://") && !ldapConfig.StartTLS && logger != nil {
		logger.Event("WARN", "LDAP connection is unencrypted", "authentication", "warning", map[string]interface{}{
			"auth.provider":  "ldap",
			"ldap.starttls":  false,
			"ldap.transport": "plain",
		})
	}
	allowedGroups := make([]*ldap.DN, 0, len(ldapConfig.AllowedGroups))
	for _, groupDN := range ldapConfig.AllowedGroups {
		parsed, err := ldap.ParseDN(groupDN)
		if err != nil {
			return nil, fmt.Errorf("LDAP_ALLOWED_GROUPS contains an invalid distinguished name")
		}
		allowedGroups = append(allowedGroups, parsed)
	}
	return &LDAPProvider{
		serverURL:         ldapConfig.URL,
		startTLS:          ldapConfig.StartTLS,
		tlsConfig:         tlsConfig,
		bindDN:            ldapConfig.BindDN,
		bindPassword:      ldapConfig.BindPassword,
		userSearchBase:    ldapConfig.UserSearchBase,
		userFilter:        ldapConfig.UserFilter,
		usernameAttribute: ldapConfig.UsernameAttribute,
		groupSearchBase:   ldapConfig.GroupSearchBase,
		groupFilter:       ldapConfig.GroupFilter,
		allowedGroups:     allowedGroups,
		sessions:          sessions,
		logger:            logger,
		dialLDAP: func(address string, options ...ldap.DialOpt) (ldapConnection, error) {
			return ldap.DialURL(address, options...)
		},
	}, nil
}

func newLDAPTLSConfig(ldapConfig config.LDAP) (*tls.Config, error) {
	parsedURL, err := url.Parse(ldapConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("LDAP_URL is invalid")
	}
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         parsedURL.Hostname(),
		InsecureSkipVerify: ldapConfig.InsecureSkipVerify,
	}
	if ldapConfig.CACertFile == "" {
		return tlsConfig, nil
	}
	certificatePEM, err := os.ReadFile(ldapConfig.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("could not read LDAP_CA_CERT_FILE")
	}
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("could not load system certificate authorities")
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if !rootCAs.AppendCertsFromPEM(certificatePEM) {
		return nil, fmt.Errorf("LDAP_CA_CERT_FILE contains no valid certificates")
	}
	tlsConfig.RootCAs = rootCAs
	return tlsConfig, nil
}

func (p *LDAPProvider) Name() string {
	return "ldap"
}

func (p *LDAPProvider) CallbackPath() string {
	return ldapLoginPath
}

func (p *LDAPProvider) RegisterRoutes(e *echo.Echo) {
	e.POST(ldapLoginPath, p.login)
}

func (p *LDAPProvider) BeginLogin(c echo.Context) error {
	pending := &PendingLogin{
		Provider:     p.Name(),
		RedirectPath: c.Request().URL.RequestURI(),
	}
	if err := p.sessions.BeginLogin(c, pending); err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start authentication")
	}
	return c.Redirect(http.StatusSeeOther, ldapLoginPage)
}

func (p *LDAPProvider) LogoutURL(string) (string, bool) {
	return "", false
}

func (p *LDAPProvider) login(c echo.Context) error {
	request, err := decodeLDAPLoginRequest(c)
	if err != nil {
		p.failure(c, "invalid_login_request")
		return err
	}
	if strings.TrimSpace(request.Username) == "" || len(request.Username) > maxLDAPUsernameSize || request.Password == "" || len(request.Password) > maxLDAPPasswordSize {
		p.failure(c, "invalid_login_request")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid login request")
	}

	pending, err := p.sessions.PendingLogin(c, p.Name())
	hasPendingLogin := err == nil
	if !hasPendingLogin {
		pending = &PendingLogin{Provider: p.Name(), RedirectPath: "/"}
	}
	identity, failure := p.authenticate(request.Username, request.Password)
	if failure != "" {
		p.failure(c, failure)
		switch failure {
		case "invalid_credentials":
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
		case "group_restriction_failed":
			return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to access this application")
		default:
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Authentication service is unavailable")
		}
	}
	if hasPendingLogin {
		if _, err := p.sessions.ConsumePendingLogin(c, p.Name()); err != nil {
			p.failure(c, "login_session_failed")
			return echo.NewHTTPError(http.StatusBadRequest, "Login session is missing or expired")
		}
	}
	redirect, err := p.sessions.CompleteLogin(c, identity, "", pending.RedirectPath)
	if err != nil {
		p.failure(c, "session_save_failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create authenticated session")
	}
	return c.JSON(http.StatusOK, map[string]string{"redirect": redirect})
}

func decodeLDAPLoginRequest(c echo.Context) (*ldapLoginRequest, error) {
	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	if err != nil || mediaType != echo.MIMEApplicationJSON {
		return nil, echo.NewHTTPError(http.StatusUnsupportedMediaType, "Content-Type must be application/json")
	}
	body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxLDAPLoginBody+1))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid login request")
	}
	if len(body) > maxLDAPLoginBody {
		return nil, echo.NewHTTPError(http.StatusRequestEntityTooLarge, "Login request is too large")
	}
	var request ldapLoginRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid login request")
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid login request")
	}
	return &request, nil
}

func (p *LDAPProvider) authenticate(username, password string) (*Identity, string) {
	conn, err := p.connect()
	if err != nil {
		return nil, "ldap_connection_failed"
	}
	defer conn.Close()

	if err := conn.Bind(p.bindDN, p.bindPassword); err != nil {
		return nil, "ldap_service_bind_failed"
	}
	filter := strings.Replace(p.userFilter, "{username}", ldap.EscapeFilter(username), 1)
	result, err := conn.Search(ldap.NewSearchRequest(
		p.userSearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		2,
		int(ldapTimeout.Seconds()),
		false,
		filter,
		[]string{p.usernameAttribute, "mail"},
		nil,
	))
	if err != nil {
		return nil, "ldap_user_search_failed"
	}
	if result == nil {
		return nil, "ldap_user_search_failed"
	}
	if len(result.Entries) == 0 {
		return nil, "invalid_credentials"
	}
	if len(result.Entries) != 1 {
		return nil, "ldap_user_search_ambiguous"
	}
	user := result.Entries[0]
	if _, err := ldap.ParseDN(user.DN); err != nil {
		return nil, "ldap_user_search_failed"
	}
	if user.GetAttributeValue(p.usernameAttribute) == "" {
		return nil, "ldap_username_missing"
	}

	groupAllowed, groupErr := p.hasAllowedGroup(conn, user.DN)
	if err := conn.Bind(user.DN, password); err != nil {
		if isLDAPInvalidCredentials(err) {
			return nil, "invalid_credentials"
		}
		return nil, "ldap_user_bind_failed"
	}
	if groupErr != nil {
		return nil, "ldap_group_search_failed"
	}
	if !groupAllowed {
		return nil, "group_restriction_failed"
	}
	return &Identity{
		Provider: p.Name(),
		Subject:  user.DN,
		Username: user.GetAttributeValue(p.usernameAttribute),
		Email:    user.GetAttributeValue("mail"),
	}, ""
}

func (p *LDAPProvider) hasAllowedGroup(conn ldapConnection, userDN string) (bool, error) {
	if len(p.allowedGroups) == 0 {
		return true, nil
	}
	filter := strings.Replace(p.groupFilter, "{user_dn}", ldap.EscapeFilter(userDN), 1)
	result, err := conn.Search(ldap.NewSearchRequest(
		p.groupSearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1000,
		int(ldapTimeout.Seconds()),
		false,
		filter,
		nil,
		nil,
	))
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, errors.New("LDAP group search returned no result")
	}
	for _, entry := range result.Entries {
		group, err := ldap.ParseDN(entry.DN)
		if err != nil {
			return false, err
		}
		for _, allowedGroup := range p.allowedGroups {
			if group.EqualFold(allowedGroup) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *LDAPProvider) connect() (ldapConnection, error) {
	dialer := &net.Dialer{Timeout: ldapTimeout}
	conn, err := p.dialLDAP(
		p.serverURL,
		ldap.DialWithDialer(dialer),
		ldap.DialWithTLSConfig(p.tlsConfig),
	)
	if err != nil {
		return nil, err
	}
	conn.SetTimeout(ldapTimeout)
	if p.startTLS {
		if err := conn.StartTLS(p.tlsConfig); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return conn, nil
}

func isLDAPInvalidCredentials(err error) bool {
	var ldapError *ldap.Error
	return errors.As(err, &ldapError) && ldapError.ResultCode == ldap.LDAPResultInvalidCredentials
}

func (p *LDAPProvider) failure(c echo.Context, errorType string) {
	if p.logger != nil {
		p.logger.AuthenticationFailure(logging.RequestID(c), p.Name(), errorType, requestSourceIP(c))
	}
}
