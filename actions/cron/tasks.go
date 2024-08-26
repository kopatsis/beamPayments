package cron

import (
	"beam_payments/models"
	"time"

	"github.com/go-co-op/gocron"
)

func ScheduledTasks() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(3).Hours().Do(models.ScheduledSubscriptionMods)
	s.StartAsync()
}
