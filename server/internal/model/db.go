package model

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	Role         string `gorm:"default:admin"`
	Enabled      bool   `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt
}

type Provider struct {
	ID               uint   `gorm:"primaryKey"`
	Name             string `gorm:"uniqueIndex"`
	OpenAIBaseURL    string `gorm:"column:openai_base_url"`
	AnthropicBaseURL string `gorm:"column:anthropic_base_url"`
	APIKey           string
	Enabled          bool   `gorm:"default:true"`
	Priority         int    `gorm:"default:0"`
	Config           string `gorm:"type:text"`
	LastSyncAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt
	Models           []ProviderModel
}

type ProviderModel struct {
	ID             uint `gorm:"primaryKey"`
	ProviderID     uint `gorm:"index"`
	ModelID        string
	DisplayName    string
	OwnedBy        string
	ContextWindow  int     `gorm:"default:0"`
	MaxOutput      int     `gorm:"default:0"`
	InputPrice     float64 `gorm:"default:0"`
	OutputPrice    float64 `gorm:"default:0"`
	SupportsVision bool    `gorm:"default:false"`
	SupportsTools  bool    `gorm:"default:true"`
	SupportsStream bool    `gorm:"default:true"`
	Metadata       string  `gorm:"type:text"`
	IsAvailable    bool    `gorm:"default:true"`
	Source         string  `gorm:"default:sync"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt
}

type Alias struct {
	ID        uint   `gorm:"primaryKey;tableName:aliases"`
	Name      string `gorm:"uniqueIndex;column:name"`
	Enabled   bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	Mappings  []AliasMapping `gorm:"foreignKey:AliasID;constraint:OnDelete:CASCADE"`
}

type AliasMapping struct {
	ID                uint `gorm:"primaryKey;tableName:alias_mappings"`
	AliasID           uint `gorm:"index"`
	ProviderID        uint `gorm:"index"`
	ProviderModelName string
	Weight            int  `gorm:"default:1"`
	Enabled           bool `gorm:"default:true"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt
	Provider          *Provider `gorm:"foreignKey:ProviderID"`
}

type Mapping struct {
	ID                uint `gorm:"primaryKey;tableName:mappings"`
	AliasID           uint `gorm:"index"`
	ProviderID        uint `gorm:"index"`
	ProviderModelName string
	Weight            int  `gorm:"default:1"`
	Enabled           bool `gorm:"default:true"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt
	Provider          *Provider `gorm:"foreignKey:ProviderID"`
}

type Key struct {
	ID        uint   `gorm:"primaryKey"`
	Key       string `gorm:"uniqueIndex"`
	Name      string
	Enabled   bool `gorm:"default:true"`
	ExpiresAt *time.Time
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
	Models    []KeyModel
}

type KeyModel struct {
	ID        uint `gorm:"primaryKey"`
	KeyID     uint `gorm:"index"`
	Model     string
	CreatedAt time.Time
}

type UsageLog struct {
	ID              uint `gorm:"primaryKey"`
	Source          string
	KeyID           uint `gorm:"index"`
	KeyName         string
	Model           string
	ProviderID      uint `gorm:"index"`
	ProviderName    string
	ActualModelID   string `gorm:"index"`
	ActualModelName string
	CallMethod      string
	TotalTokens     int `gorm:"default:0"`
	LatencyMs       int `gorm:"default:0"`
	Status          string
	ErrorMsg        string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"index"`
}

func (u *UsageLog) String() string {
	return fmt.Sprintf("[%s] %s calling model %s, provider:(%s/%s), method:%s, tokens:%d, time:%dms, status:%s",
		u.Source, u.KeyName, u.Model, u.ProviderName, u.ActualModelName, u.CallMethod, u.TotalTokens, u.LatencyMs, u.Status,
	)
}

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

	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&User{},
		&Provider{},
		&ProviderModel{},
		&Alias{},
		&AliasMapping{},
		&Key{},
		&KeyModel{},
		&UsageLog{},
	)
}

func GetDB() *gorm.DB {
	return DB
}
