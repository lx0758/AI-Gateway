package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string         `gorm:"size:256;not null" json:"-"`
	Role         string         `gorm:"size:32;default:admin" json:"role"`
	Enabled      bool           `gorm:"default:true" json:"enabled"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Provider struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	Name         string          `gorm:"uniqueIndex;size:128;not null" json:"name"`
	APIType      string          `gorm:"size:32;not null" json:"api_type"`
	BaseURL      string          `gorm:"size:512;not null" json:"base_url"`
	APIKey       string          `gorm:"size:256" json:"-"`
	APIKeyMasked string          `gorm:"-" json:"api_key_masked"`
	Enabled      bool            `gorm:"default:true" json:"enabled"`
	Priority     int             `gorm:"default:0" json:"priority"`
	Config       string          `gorm:"type:text" json:"config"`
	LastSyncAt   *time.Time      `json:"last_sync_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
	Models       []ProviderModel `gorm:"foreignKey:ProviderID" json:"models,omitempty"`
}

type ProviderModel struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProviderID     uint           `gorm:"index;not null" json:"provider_id"`
	ModelID        string         `gorm:"size:128;not null" json:"model_id"`
	DisplayName    string         `gorm:"size:128" json:"display_name"`
	OwnedBy        string         `gorm:"size:64" json:"owned_by"`
	ContextWindow  int            `gorm:"default:0" json:"context_window"`
	MaxOutput      int            `gorm:"default:0" json:"max_output"`
	InputPrice     float64        `gorm:"default:0" json:"input_price"`
	OutputPrice    float64        `gorm:"default:0" json:"output_price"`
	SupportsVision bool           `gorm:"default:false" json:"supports_vision"`
	SupportsTools  bool           `gorm:"default:true" json:"supports_tools"`
	SupportsStream bool           `gorm:"default:true" json:"supports_stream"`
	Metadata       string         `gorm:"type:text" json:"metadata"`
	IsAvailable    bool           `gorm:"default:true" json:"is_available"`
	Source         string         `gorm:"size:32;default:sync" json:"source"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type ModelMapping struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Alias             string         `gorm:"index;size:128;not null" json:"alias"`
	ProviderID        uint           `gorm:"index;not null" json:"provider_id"`
	ProviderModelName string         `gorm:"index;size:128;not null" json:"provider_model_name"`
	Enabled           bool           `gorm:"default:true" json:"enabled"`
	Weight            int            `gorm:"default:1" json:"weight"`
	CreatedAt         time.Time      `json:"created_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	Provider          *Provider      `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

type APIKey struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Key           string         `gorm:"uniqueIndex;size:64;not null" json:"key"`
	Name          string         `gorm:"size:128" json:"name"`
	AllowedModels string         `gorm:"type:text" json:"allowed_models"`
	RateLimit     int            `gorm:"default:0" json:"rate_limit"`
	Quota         int64          `gorm:"default:0" json:"quota"`
	UsedQuota     int64          `gorm:"default:0" json:"used_quota"`
	UsedCount     int64          `gorm:"default:0" json:"used_count"`
	ExpiresAt     *time.Time     `json:"expires_at"`
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Models        []APIKeyModel  `gorm:"foreignKey:APIKeyID" json:"models,omitempty"`
}

type APIKeyModel struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	APIKeyID   uint      `gorm:"index;not null" json:"api_key_id"`
	ModelAlias string    `gorm:"size:128;not null" json:"model_alias"`
	CreatedAt  time.Time `json:"created_at"`
}

type UsageLog struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	APIKeyID         uint      `gorm:"index" json:"api_key_id"`
	ProviderID       uint      `gorm:"index" json:"provider_id"`
	Model            string    `gorm:"size:128;not null" json:"model"`
	ActualModel      string    `gorm:"size:128" json:"actual_model"`
	PromptTokens     int       `gorm:"default:0" json:"prompt_tokens"`
	CompletionTokens int       `gorm:"default:0" json:"completion_tokens"`
	LatencyMs        int       `gorm:"default:0" json:"latency_ms"`
	Status           string    `gorm:"size:32;not null" json:"status"`
	ErrorMsg         string    `gorm:"type:text" json:"error_msg"`
	CreatedAt        time.Time `gorm:"index" json:"created_at"`
}
