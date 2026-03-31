package model

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	dsn := dbPath + "?_loc=auto"
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	if err := autoMigrate(); err != nil {
		return err
	}

	return migrateAllowedModels()
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&User{},
		&Provider{},
		&ProviderModel{},
		&ModelMapping{},
		&Key{},
		&KeyModel{},
		&UsageLog{},
	)
}

func migrateAllowedModels() error {
	var keys []Key
	if err := DB.Where("allowed_models != ? AND allowed_models != ''", "[]").Find(&keys).Error; err != nil {
		return err
	}

	for _, key := range keys {
		if key.AllowedModels == "" {
			continue
		}

		var models []string
		if err := json.Unmarshal([]byte(key.AllowedModels), &models); err != nil {
			continue
		}

		for _, alias := range models {
			var existing KeyModel
			if err := DB.Where("key_id = ? AND model_alias = ?", key.ID, alias).First(&existing).Error; err == nil {
				continue
			}

			akm := KeyModel{
				KeyID:      key.ID,
				ModelAlias: alias,
			}
			DB.Create(&akm)
		}

		DB.Model(&key).Update("allowed_models", "[]")
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
