package actions

import (
	"beam_payments/models"
	"errors"

	"github.com/gobuffalo/buffalo"
)

func ExternalGetHandler(c buffalo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.Error(400, errors.New("no param provided"))
	}

	sub, exists, err := models.GetSubscription(id)
	if err != nil {
		return c.Error(400, err)
	}

	if !exists {
		return c.Render(200, r.JSON(map[string]any{"id": "", "paying": false}))
	}

	if !sub.Paying {
		return c.Render(200, r.JSON(map[string]any{"id": sub.SubscriptionID, "paying": false}))
	}

	return c.Render(200, r.JSON(map[string]any{"id": sub.SubscriptionID, "paying": true}))
}
