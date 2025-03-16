package service

import (
	"github.com/yourusername/project/pkg/databaseManager"

	"gorm.io/gorm"
)

func ProvideGormDB(dbManager databaseManager.DatabaseManager) *gorm.DB {
	return dbManager.GetDB()
}
