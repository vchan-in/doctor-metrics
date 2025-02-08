package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"vchan.in/docker-health/types"
)

func GetRoot(c echo.Context, version string) error {
	/*
		GetRoot is a handler function that returns the root endpoint response.
		It returns a JSON response with a success status and a message.
	*/
	return c.JSON(http.StatusOK, types.APIResponse{
		Status:  "success",
		Message: "Docker Health API v" + version,
	})
}
