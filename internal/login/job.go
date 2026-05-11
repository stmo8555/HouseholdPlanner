package login

import (
	"context"
	"github.com/robfig/cron/v3"
)

func RunCleanup(ctx context.Context, service *Service) {
	c := cron.New()
	c.AddFunc("@every 1h", func() {
		err := service.RemoveExpiredSessions(ctx)
		if err != nil {
			panic(err)
		}
	})
	c.Start()
}
