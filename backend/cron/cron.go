package cron

import (
	"context"
	"news-swipe/backend/utils"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func CreateCron(ctx context.Context, db *gorm.DB) {
	err := FilterLinked(db)
	if err != nil {
		utils.Log(utils.Database, err)
	}
	if err := utils.GenerateKeywordsFromArticles(db); err != nil {
		utils.Log(utils.Database, "Keyword generation failed", "error", err)
	}
	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))

	_, err = c.AddFunc("*/15 * * * *", func() {
		err := FilterLinked(db)
		if err != nil {
			utils.Log(utils.Database, err)
		}
	})
	if err != nil {
		utils.Log(utils.Cron, err)
	}
	_, err = c.AddFunc("0 4 * * *", func() {
		if err := CleanupOldArticles(db, 7); err != nil {
			utils.Log(utils.Database, "Article cleanup failed: "+err.Error())
		}
	})
	if err != nil {
		utils.Log(utils.Cron, err)
	}
	_, err = c.AddFunc("0 5 * * *", func() { // 5 AM daily
		if err := utils.GenerateKeywordsFromArticles(db); err != nil {
			utils.Log(utils.Database, "Keyword generation failed", "error", err)
		}
	})
	if err != nil {
		utils.Log(utils.Cron, err)
	}

	c.Start()
	utils.Log(utils.Cron, "CronJob is started")
	<-ctx.Done()
	utils.Log(utils.Cron, "CronJob is shutedown")
	c.Stop()
}
