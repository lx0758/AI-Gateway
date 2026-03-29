package model

import (
	"golang.org/x/crypto/bcrypt"
)

func InitDefaultAdmin(username, password string) error {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		Enabled:      true,
	}

	return DB.Create(&user).Error
}
