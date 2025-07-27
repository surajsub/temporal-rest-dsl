package handlers

import (
	"github.com/labstack/echo/v4"
	"go.temporal.io/sdk/client"
	"net/http"
)

func RegisterRoutes(e *echo.Echo, getClient func() client.Client) {
	e.POST("/v1/provision", func(c echo.Context) error {
		client := getClient()
		if client == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Temporal client not available"})
		}
		return SubmitWorkflowHandler(c, client)
	})

	e.GET("/v1/status/:submission_id", func(c echo.Context) error {
		client := getClient()
		if client == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Temporal client not available"})
		}
		return GetSubmissionIDStatus(c, client)
	})

	e.GET("/v1/newstatus/:workflow_id", func(c echo.Context) error {
		client := getClient()
		if client == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Temporal client not available"})
		}
		return GetWorkflowActivityHistoryHandler(c, client)
	})
	e.POST("/v1/retry", func(c echo.Context) error {
		client := getClient()
		if client == nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Temporal client not available"})
		}
		return NewSendSignalHandler(c, client)
	})
}
