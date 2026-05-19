package todo

import (
	"context"
	"github.com/robfig/cron/v3"
)

func RunCleanup(ctx context.Context, service *Service) {
	c := cron.New()
	c.AddFunc("@every 1h", func() {
		service.RemoveOldCompleted(context.Background())
	})
	c.Start()
}

func ScheduleRepeats(ctx context.Context, service *Service) {
	c := cron.New()
	c.AddFunc("@every 12h", func() {
		service.ScheduleRepeats(context.Background())
	})
	c.Start()
}
