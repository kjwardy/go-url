package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

// SlackPayload received from slack
type SlackPayload struct {
	Token          string `form:"token"`
	TeamID         string `form:"team_id"`
	TeamDomain     string `form:"team_domain"`
	EnterpriseID   string `form:"enterprise_id"`
	EnterpriseName string `form:"enterprise_name"`
	ChannelID      string `form:"channel_id"`
	ChannelName    string `form:"channel_name"`
	UserID         string `form:"user_id"`
	UserName       string `form:"user_name"`
	Command        string `form:"command"`
	Text           string `form:"text"`
	ResponseURL    string `form:"response_url"`
	TriggerID      string `form:"trigger_id"`
}

type SlackAttachment struct {
	Title  string `json:"title"`
	Text   string `json:"text"`
	Author string `json:"author_text"`
	Footer string `json:"footer"`
}

type SlackResponse struct {
	ResponseType string            `json:"response_type"`
	Text         string            `json:"text"`
	Attachments  []SlackAttachment `json:"attachments"`
	Parse        string            `json:"parse"`
}

type SlackSecretsVerifier struct {
	rawBody       []byte
	slackSig      string
	timeStamp     string
	signingSecret string
}

var (
	errSlackBodyRead  = errors.New("slack request body could not be read")
	errSlackSignature = errors.New("slack request signature is invalid")
	errSlackTeam      = errors.New("slack team is not allowed")
)

func verifySignedSecret(v *SlackSecretsVerifier) bool {
	message := fmt.Sprintf("v0:%v:%v", v.timeStamp, string(v.rawBody))
	expectedMAC := v.slackSig

	mac := hmac.New(sha256.New, []byte(v.signingSecret))
	mac.Write([]byte(message))
	actualSignature := "v0=" + string(hex.EncodeToString(mac.Sum(nil)))

	if actualSignature == expectedMAC {
		return true
	}

	return false
}

// Optionally only allow requests from specified slack team
func verifySlackTeam(payload *SlackPayload) error {
	appConfig := config.GetConfig()
	teamID := appConfig.Slack.TeamID
	if teamID != "" && teamID != payload.TeamID {
		return errSlackTeam
	}
	return nil
}

// verify request has been signed by slack
func verifySlackRequest(r *http.Request) error {
	buf, err := ioutil.ReadAll(r.Body)
	body := ioutil.NopCloser(bytes.NewBuffer(buf))
	if err != nil {
		return errSlackBodyRead
	}
	r.Body = body
	appConfig := config.GetConfig()

	v := SlackSecretsVerifier{
		rawBody:       buf,
		slackSig:      r.Header.Get("X-Slack-Signature"),
		timeStamp:     r.Header.Get("X-Slack-Request-Timestamp"),
		signingSecret: appConfig.Slack.SigningSecret,
	}

	if !verifySignedSecret(&v) {
		return errSlackSignature
	}
	return nil
}

func (h *Handler) logSlackCommand(c echo.Context, level, message, result, errorType string) {
	if h.Logger == nil {
		return
	}
	fields := map[string]interface{}{
		"http.request.id": logging.RequestID(c),
		"slack.operation": "command",
	}
	if errorType != "" {
		fields["error.type"] = errorType
	}
	h.Logger.Event(level, message, "slack.command", result, fields)
}

// SlackCommand responds to slack go command
func (h *Handler) SlackCommand(c echo.Context) error {
	// Reading request breaks rest
	if err := verifySlackRequest(c.Request()); err != nil {
		errorType := "invalid_signature"
		if errors.Is(err, errSlackBodyRead) {
			errorType = "request_body_read_error"
		}
		h.logSlackCommand(c, "WARN", "Slack command authentication failed", "failure", errorType)
		return c.JSON(http.StatusUnauthorized, "Unauthorized")
	}
	response := new(SlackResponse)
	response.ResponseType = "ephemeral"
	//urlModel.Find()
	payload := new(SlackPayload)
	if err := c.Bind(payload); err != nil {
		h.logSlackCommand(c, "WARN", "Slack command payload is invalid", "failure", "invalid_payload")
		response.Text = "An error occurred while processing your request"
		return c.JSON(http.StatusOK, response)
	}

	if err := verifySlackTeam(payload); err != nil {
		h.logSlackCommand(c, "WARN", "Slack command team is not allowed", "failure", "team_not_allowed")
		return c.JSON(http.StatusUnauthorized, "Unauthorized")
	}

	key := payload.Text
	if ValidateKeyPath(key) == false {
		h.logSlackCommand(c, "WARN", "Slack command key is invalid", "failure", "invalid_key")
		response.Text = "The URL key you provided is an invalid format. You can only provide one key at a time, e.g. `/go help`"
		return c.JSON(http.StatusOK, response)
	}
	appURI := config.GetConfig().AppURI
	urls, err := urlModel.GetUrlsFromKeys([]string{key})
	if err != nil {
		h.logSlackCommand(c, "ERROR", "Slack command URL lookup failed", "failure", "database_error")
		response.Text = fmt.Sprintf("The URL key you provided was not found. Why not add it? %s/go/%s", appURI, key)
		return c.JSON(http.StatusOK, response)
	}
	if len(urls) == 0 {
		h.logSlackCommand(c, "INFO", "Slack command URL was not found", "failure", "url_not_found")
		response.Text = fmt.Sprintf("The URL key you provided was not found. Why not add it? %s/go/%s", appURI, key)
		return c.JSON(http.StatusOK, response)
	}
	url := urls[0]
	response.Text = fmt.Sprintf("%s/%s", appURI, key)
	if url.URL == "" {
		aliasURLs := make([]string, len(url.Alias))
		for i, alias := range url.Alias {
			aliasURLs[i] = fmt.Sprintf("<%s/%s|%s>", appURI, alias, alias)
		}
		response.Attachments = []SlackAttachment{{
			Footer: fmt.Sprintf("Points to the alias `%s`", strings.Join(aliasURLs, "`, `")),
		}}
	} else {
		response.Attachments = []SlackAttachment{{
			Footer: url.URL,
		}}
	}
	h.logSlackCommand(c, "INFO", "Slack command completed", "success", "")
	return c.JSON(http.StatusOK, response)
}
