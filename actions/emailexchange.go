package actions

import (
	"beam_payments/redis"
	"net/http"
	"os"

	"github.com/gobuffalo/buffalo"
)

func AddExchange(c buffalo.Context) error {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.Render(http.StatusBadRequest, r.JSON(map[string]string{"error": "Invalid JSON"}))
	}

	if c.Request().Header.Get("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
		return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{"error": "Unauthorized"}))
	}

	id, err := redis.AddEmail(req.Email)
	if err != nil {
		return c.Render(http.StatusInternalServerError, r.JSON(map[string]string{"error": "Failed to store data"}))
	}

	return c.Render(http.StatusOK, r.JSON(map[string]string{"id": id}))
}

func GetExchange(c buffalo.Context) error {
	if c.Request().Header.Get("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
		return c.Render(http.StatusUnauthorized, r.JSON(map[string]string{"error": "Unauthorized"}))
	}

	id := c.Param("id")

	email, err := redis.GetAndDeleteEmail(id)
	if err != nil {
		return c.Render(http.StatusNotFound, r.JSON(map[string]string{"error": "Key not found or failed to retrieve"}))
	}

	return c.Render(http.StatusOK, r.JSON(map[string]string{"email": email}))
}
