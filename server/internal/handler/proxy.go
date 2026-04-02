package handler

import (
	"fmt"

	modelPkg "ai-gateway/internal/model"
)

func VerifyKeyID(keyID any, model string) error {
	if _, ok := keyID.(uint); !ok {
		return fmt.Errorf("invalid key")
	}
	validKeyID := keyID.(uint)
	var count int64
	modelPkg.DB.Model(&modelPkg.KeyModel{}).Where("key_id = ?", validKeyID).Count(&count)
	if count == 0 {
		return nil
	}

	var alias modelPkg.Alias
	if err := modelPkg.DB.Where("name = ?", model).First(&alias).Error; err != nil {
		return fmt.Errorf("model not allowed for this API key")
	}

	var modelCount int64
	modelPkg.DB.Model(&modelPkg.KeyModel{}).Where("key_id = ? AND alias_id = ?", validKeyID, alias.ID).Count(&modelCount)
	if modelCount == 0 {
		return fmt.Errorf("model not allowed for this API key")
	}
	return nil
}
