package cron

import (
	"time"

	"github.com/go-co-op/gocron"
)

func ScheduledTasks() {
	s := gocron.NewScheduler(time.UTC)
	// s.Every(1).Hours().Do(models.ScheduledSubscriptionMods)
	s.StartAsync()
}
