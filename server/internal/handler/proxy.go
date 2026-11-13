package handler

import (
	"fmt"

	modelPkg "ai-gateway/internal/model"
)

// VerifyKeyID 按模型白名单开关两段式校验（设计 D2）：
// 开关关闭 = 直接放行（不查关联表）；开关开启 = 仅允许已勾选的模型。
func VerifyKeyID(keyID any, model string) error {
	if _, ok := keyID.(uint); !ok {
		return fmt.Errorf("invalid key")
	}
	validKeyID := keyID.(uint)

	var key modelPkg.Key
	if err := modelPkg.DB.First(&key, validKeyID).Error; err != nil {
		return fmt.Errorf("invalid key")
	}
	if !key.ModelsWhitelist {
		return nil
	}

	var m modelPkg.Model
	if err := modelPkg.DB.Where("name = ?", model).First(&m).Error; err != nil {
		return fmt.Errorf("model not allowed for this API key")
	}

	var modelCount int64
	modelPkg.DB.Model(&modelPkg.KeyModel{}).Where("key_id = ? AND model_id = ?", validKeyID, m.ID).Count(&modelCount)
	if modelCount == 0 {
		return fmt.Errorf("model not allowed for this API key")
	}
	return nil
}
