package cron

import (
	"log"

	"github.com/robfig/cron/v3"
	db "github.com/suprimkhatri77/turgorepo/api/internal/database/generated"
)

func CronExample(queries *db.Queries) {
	c := cron.New()

	_, err := c.AddFunc("0 0 * * *", func() {
		// TODO: Implement a job
	})
	if err != nil {
		log.Fatalf("cron: failed to schedule expire subscriptions job: %v", err)
	}

	c.Start()
}
