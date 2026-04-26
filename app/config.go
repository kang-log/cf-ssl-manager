package app

import (
	"os"
	"path/filepath"
)

func GetCertPath() string {
	var s Setting
	if gormDB != nil {
		if err := gormDB.Where("key = ?", "certPath").First(&s).Error; err == nil && s.Value != "" {
			return s.Value
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "cf-ssl-certs")
}

func GetSetting(key string) string {
	if gormDB == nil {
		return ""
	}
	var s Setting
	if err := gormDB.Where("key = ?", key).First(&s).Error; err != nil {
		return ""
	}
	return s.Value
}

func SaveSetting(key, value string) error {
	if gormDB == nil {
		return nil
	}
	return gormDB.Save(&Setting{Key: key, Value: value}).Error
}
