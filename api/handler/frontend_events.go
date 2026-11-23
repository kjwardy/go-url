package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

const (
	frontendEventVersion = 1
	frontendEventLimit   = 20
	frontendEventWindow  = time.Minute
	maxFrontendEventSize = 1024
	maxTrackedEventIPs   = 10000
)

var frontendComponents = map[string]bool{
	"home":        true,
	"search":      true,
	"history":     true,
	"metrics":     true,
	"most_wanted": true,
	"edit_url":    true,
	"delete_url":  true,
}

var frontendOperations = map[string]string{
	"url.create": "edit_url",
	"url.update": "edit_url",
	"url.delete": "delete_url",
}

type frontendEvent struct {
	Version   int    `json:"version"`
	Type      string `json:"type"`
	Component string `json:"component,omitempty"`
	Code      string `json:"code,omitempty"`
	Operation string `json:"operation,omitempty"`
	Outcome   string `json:"outcome,omitempty"`
}

type frontendEventRate struct {
	started time.Time
	count   int
}

type FrontendEventLimiter struct {
	mu          sync.Mutex
	clients     map[string]frontendEventRate
	lastCleanup time.Time
	now         func() time.Time
}

func NewFrontendEventLimiter() *FrontendEventLimiter {
	return &FrontendEventLimiter{
		clients: make(map[string]frontendEventRate),
		now:     time.Now,
	}
}

func (l *FrontendEventLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if now.Sub(l.lastCleanup) >= frontendEventWindow {
		for clientIP, entry := range l.clients {
			if now.Sub(entry.started) >= frontendEventWindow {
				delete(l.clients, clientIP)
			}
		}
		l.lastCleanup = now
	}

	entry, exists := l.clients[ip]
	if !exists {
		if len(l.clients) >= maxTrackedEventIPs {
			return false
		}
		entry = frontendEventRate{started: now}
	}
	if now.Sub(entry.started) >= frontendEventWindow {
		entry = frontendEventRate{started: now}
	}
	if entry.count >= frontendEventLimit {
		l.clients[ip] = entry
		return false
	}
	entry.count++
	l.clients[ip] = entry
	return true
}

func (h *Handler) ReportFrontendEvent(c echo.Context) error {
	if h.FrontendEventLimiter == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Frontend event reporting is unavailable")
	}

	ip := remoteIP(c.Request().RemoteAddr)
	if !h.FrontendEventLimiter.allow(ip) {
		return echo.NewHTTPError(http.StatusTooManyRequests, "Frontend event rate limit exceeded")
	}

	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	if err != nil || mediaType != echo.MIMEApplicationJSON {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Content-Type must be application/json")
	}

	body, err := io.ReadAll(io.LimitReader(c.Request().Body, maxFrontendEventSize+1))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid frontend event")
	}
	if len(body) > maxFrontendEventSize {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "Frontend event payload too large")
	}

	var event frontendEvent
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid frontend event")
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid frontend event")
	}
	if event.Version != frontendEventVersion {
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported frontend event version")
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || !frontendEventFieldsMatch(event.Type, fields) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid frontend event")
	}

	action, component, level, outcome, errorType, message, eventType, valid := normalizeFrontendEvent(event)
	if !valid {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid frontend event")
	}

	if h.Logger == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Frontend event reporting is unavailable")
	}
	attributes := map[string]interface{}{
		"event.category":     []string{"web"},
		"event.type":         []string{eventType},
		"frontend.component": component,
		"http.request.id":    logging.RequestID(c),
	}
	if ip != "" {
		attributes["source.ip"] = ip
	}
	if errorType != "" {
		attributes["error.type"] = errorType
	}
	if h.Auth != nil {
		if identity, err := h.Auth.Identity(c); err == nil {
			if identity.Username != "" {
				attributes["user.name"] = identity.Username
			}
			if identity.Provider != "" {
				attributes["auth.provider"] = identity.Provider
			}
		}
	}
	h.Logger.Event(level, message, action, outcome, attributes)
	return c.NoContent(http.StatusAccepted)
}

func frontendEventFieldsMatch(eventType string, fields map[string]json.RawMessage) bool {
	allowed := map[string]bool{"version": true, "type": true}
	switch eventType {
	case "error":
		allowed["component"] = true
		allowed["code"] = true
	case "operation":
		allowed["operation"] = true
		allowed["outcome"] = true
	default:
		return false
	}
	if len(fields) != len(allowed) {
		return false
	}
	for field := range fields {
		if !allowed[field] {
			return false
		}
	}
	return true
}

func normalizeFrontendEvent(event frontendEvent) (action, component, level, outcome, errorType, message, eventType string, valid bool) {
	switch event.Type {
	case "error":
		if event.Operation != "" || event.Outcome != "" || !frontendComponents[event.Component] || event.Code != "api_request_failed" {
			return "", "", "", "", "", "", "", false
		}
		return "frontend.error", event.Component, "ERROR", "failure", event.Code, "Frontend error reported", "error", true
	case "operation":
		component, ok := frontendOperations[event.Operation]
		if !ok || event.Component != "" || event.Code != "" || (event.Outcome != "success" && event.Outcome != "failure") {
			return "", "", "", "", "", "", "", false
		}
		return "frontend." + event.Operation, component, frontendOperationLevel(event.Outcome), event.Outcome, "", "Frontend operation completed", "info", true
	default:
		return "", "", "", "", "", "", "", false
	}
}

func frontendOperationLevel(outcome string) string {
	if outcome == "failure" {
		return "WARN"
	}
	return "INFO"
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = strings.Trim(remoteAddr, "[]")
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return ""
}
