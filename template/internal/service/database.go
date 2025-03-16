package service

import (
	"passontw-slot-game/pkg/databaseManager"

	"gorm.io/gorm"
)

func ProvideGormDB(dbManager databaseManager.DatabaseManager) *gorm.DB {
	return dbManager.GetDB()
}
