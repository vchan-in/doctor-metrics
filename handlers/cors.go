package handlers

import (
	"os"

	"github.com/labstack/echo/v4"
)

func HandleCORS(next echo.HandlerFunc) echo.HandlerFunc {
	/*
		HandleCORS is a middleware function that adds the necessary headers to enable CORS.
		It sets the Access-Control-Allow-Origin header to allow all origins.
		It sets the Access-Control-Allow-Methods header to allow GET, POST, PUT, DELETE, OPTIONS.
		It sets the Access-Control-Allow-Headers header to allow Content-Type, Authorization.
		It sets the Access-Control-Expose-Headers header to allow Content-Length, Content-Range.
		It sets the Access-Control-Allow-Credentials header to true.
	*/
	return func(c echo.Context) error {
		origin := os.Getenv("DH_CORS_ORIGIN")
		if origin == "" {
			origin = "*"
		}

		c.Response().Header().Set("Access-Control-Allow-Origin", origin)
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		return next(c)
	}
}
