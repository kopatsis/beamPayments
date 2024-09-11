package actions

import (
	"beam_payments/models"
	"beam_payments/redis"
	"errors"
	"os"

	"github.com/gobuffalo/buffalo"
)

func ExternalGetHandler(c buffalo.Context) error {

	shouldBe := os.Getenv("CHECK_PASSCODE")
	if shouldBe == "" {
		return c.Error(500, errors.New("no passcode exists on backend"))
	}

	passcode := c.Request().Header.Get("X-Passcode-ID")
	if passcode == "" {
		return c.Error(400, errors.New("no passcode provided"))
	} else if passcode != shouldBe {
		return c.Error(400, errors.New("incorrect passcode provided"))
	}

	id := c.Param("id")
	if id == "" {
		return c.Error(400, errors.New("no param provided"))
	}

	paying, err := redis.GetSub(id)
	if err != nil {
		sub, blank, err := models.GetSubscription(id)
		if err != nil {
			return c.Error(400, err)
		}

		if blank {
			return c.Render(200, r.JSON(map[string]any{"id": "", "paying": false}))
		}

		if !sub.Paying {
			return c.Render(200, r.JSON(map[string]any{"id": sub.SubscriptionID, "paying": false}))
		}
	}

	return c.Render(200, r.JSON(map[string]any{"paying": paying}))
}
