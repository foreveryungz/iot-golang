package delivery

import (
	"log"

	"github.com/robfig/cron/v3"
)

func StartScheduler(service Service) error {
	c := cron.New(cron.WithSeconds())

	// send pending every 30 seconds
	if _, err := c.AddFunc("*/30 * * * * *", func() {
		log.Println("Running pending delivery scheduler")
		if _, err := service.ProcessPending(); err != nil {
			log.Println(err)
		}
	}); err != nil {
		return err
	}

	// retry failed every 1 minute
	if _, err := c.AddFunc("0 * * * * *", func() {
		log.Println("Running retry scheduler")
		if _, err := service.ProcessRetryable(); err != nil {
			log.Println(err)
		}
	}); err != nil {
		return err
	}

	c.Start()
	return nil
}
