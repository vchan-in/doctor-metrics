package handlers

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// Basic authentication middleware
func HandleAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	/*
		HandleAuthMiddleware is a middleware function that checks if the provided credentials are valid.
		It retrieves the environment variables, and checks if the provided credentials match the environment variables.
		If the credentials are valid, the request is passed to the next handler.
		If the credentials are invalid, an HTTP 401 Unauthorized error is returned.
	*/
	return func(c echo.Context) error {
		// Retrieve the environment variables
		username := os.Getenv("DH_USERNAME")
		password := os.Getenv("DH_PASSWORD")

		// Check if the provided credentials are valid
		auth := c.Request().Header.Get("Authorization")
		if auth == "" { // Check if the Authorization header is present
			return echo.ErrUnauthorized
		}

		payload, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		if err != nil {
			return echo.ErrUnauthorized
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != username || pair[1] != password {
			return echo.ErrUnauthorized
		}

		return next(c)
	}
}
