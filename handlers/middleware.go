package handlers

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
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
		username := os.Getenv("DM_USERNAME")
		password := os.Getenv("DM_PASSWORD")

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

func FilterIP(next echo.HandlerFunc) echo.HandlerFunc {
	/*
		FilterIP is a middleware function that only allows requests from specific IP addresses or CIDR ranges listed in the DM_ALLOWED_IPS environment variable.
		If the request is from localhost, an HTTP 401 Unauthorized error is returned.
	*/
	return func(c echo.Context) error {
		// Retrieve the environment variable
		allowedIPs := os.Getenv("DM_ALLOWED_IPS")
		if allowedIPs == "" {
			slog.Error("DM_ALLOWED_IPS environment variable not set")
			return echo.ErrUnauthorized
		}

		// Get the client's IP address
		clientIP := c.RealIP()

		// Additional checks for headers
		if clientIP == "" {
			clientIP = c.Request().Header.Get("X-Forwarded-For")
		}
		if clientIP == "" {
			clientIP = c.Request().Header.Get("X-Real-IP")
		}

		// Check if the client's IP address is in the allowed IPs list or CIDR ranges
		for _, ip := range strings.Split(allowedIPs, ",") {
			if ip == clientIP {
				return next(c)
			}
			_, cidr, err := net.ParseCIDR(ip)
			if err == nil && cidr.Contains(net.ParseIP(clientIP)) {
				return next(c)
			}
		}

		slog.Error(fmt.Sprintf("Unauthorized client IP: %s", clientIP))
		return echo.ErrUnauthorized
	}
}
