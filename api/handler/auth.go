package handler

import (
	"net/http"

	"github.com/labstack/echo"
)

func (h *Handler) auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if h.Auth == nil {
			return echo.NewHTTPError(http.StatusForbidden, "Invalid auth config")
		}
		return h.Auth.Authenticate(next, c)
	}
}
