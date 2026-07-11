package household

import (
	"context"

	"github.com/robfig/cron/v3"
)

func RunCleanup(ctx context.Context, service *Service) {
	c := cron.New()
	c.AddFunc("@every 1h", func() {
		if err := service.RemoveExpiredInvites(ctx); err != nil {
			panic(err)
		}
	})
	c.Start()
}
