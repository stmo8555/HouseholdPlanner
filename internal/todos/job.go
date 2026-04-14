package todos

import (
	"context"
	"github.com/robfig/cron/v3"
)

func RunCleanup(ctx context.Context, service *Service) {
	c := cron.New()
	c.AddFunc("0 * * * *", func() {
		service.RemoveOldCompleted(context.Background())
	})
	c.Start()
}
