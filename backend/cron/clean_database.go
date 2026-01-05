package cron

import (
	"fmt"
	"time"

	"news-swipe/backend/graph/model"
	"news-swipe/backend/utils"

	"gorm.io/gorm"
)

// CleanupOldArticles deletes articles older than the specified number of days.
// By default, it removes articles where published_at is older than 7 days.
// It uses soft delete by default if your model has DeletedAt; use permanent delete if needed.
func CleanupOldArticles(db *gorm.DB, olderThanDays int) error {
	if olderThanDays <= 0 {
		olderThanDays = 7 // default: 7 days
	}

	cutoffTime := time.Now().AddDate(0, 0, -olderThanDays)

	utils.Log(utils.Database, fmt.Sprintf("Starting cleanup: removing articles older than %d days (before %s)", olderThanDays, cutoffTime.Format("2006-01-02")))

	var deleteResult *gorm.DB

	// Option 1: Soft delete (recommended if you're using gorm.DeletedAt)
	// This marks records as deleted but keeps them in DB (good for recovery/auditing)
	deleteResult = db.Where("published_at < ?", cutoffTime).
		Delete(&model.Article{})

	// Option 2: Permanent delete (uncomment if you want to fully remove records)
	// deleteResult = db.Unscoped().
	// 	Where("published_at < ?", cutoffTime).
	// 	Delete(&model.Article{})

	if deleteResult.Error != nil {
		errStr, _ := utils.HandleGormError(deleteResult.Error)
		return fmt.Errorf("failed to delete old articles: %s", errStr)
	}

	rowsAffected := deleteResult.RowsAffected
	if rowsAffected > 0 {
		utils.Log(utils.Database, fmt.Sprintf("Cleanup complete: %d old article(s) deleted", rowsAffected))
	} else {
		utils.Log(utils.Database, "Cleanup complete: no old articles found")
	}

	return nil
}
