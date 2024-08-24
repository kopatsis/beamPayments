package actions

import (
	"beam_payments/models"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gofrs/uuid"
)

func HomeHandler(c buffalo.Context) error {

	subscription := models.Subscription{
		ID:             uuid.Must(uuid.NewV4()),
		UserID:         "test-user",
		SubscriptionID: "test-subscription",
		Processing:     true,
		Ending:         false,
		Paying:         true,
		EndDate:        time.Now().AddDate(0, 1, 0),
		ExpiresDate:    time.Now().AddDate(0, 2, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := models.DB.Create(&subscription); err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}
