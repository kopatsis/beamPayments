package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
)

// Subscription is used by pop to map your subscriptions database table to your go code.
type Subscription struct {
	ID             int       `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	SubscriptionID string    `json:"subscription_id" db:"subscription_id"`
	Processing     bool      `json:"processing" db:"processing"`
	Ending         bool      `json:"ending" db:"ending"`
	Paying         bool      `json:"paying" db:"paying"`
	EndDate        time.Time `json:"end_date" db:"end_date"`
	ExpiresDate    time.Time `json:"expires_date" db:"expires_date"`
	Archived       bool      `json:"archived" db:"archived"`
	ArchivedDate   time.Time `json:"archived_date" db:"archived_date"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (s Subscription) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Subscriptions is not required by pop and may be deleted
type Subscriptions []Subscription

// String is not required by pop and may be deleted
func (s Subscriptions) String() string {
	js, _ := json.Marshal(s)
	return string(js)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *Subscription) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: s.ID, Name: "ID"},
		&validators.StringIsPresent{Field: s.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: s.SubscriptionID, Name: "SubscriptionID"},
		&validators.TimeIsPresent{Field: s.EndDate, Name: "EndDate"},
		&validators.TimeIsPresent{Field: s.ExpiresDate, Name: "ExpiresDate"},
		&validators.TimeIsPresent{Field: s.ArchivedDate, Name: "ArchivedDate"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (s *Subscription) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (s *Subscription) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
