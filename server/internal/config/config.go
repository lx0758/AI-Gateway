package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Debug    DebugConfig
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Pprof    PprofConfig
}

type ServerConfig struct {
	Port           int
	TrustedProxies []string
	Session        SessionConfig
}

type DebugConfig struct {
	Gin      bool
	Gorm     bool
	Provider bool
	MCP      bool
}

type DatabaseConfig struct {
	Type     string
	Path     string
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
	Pool     PoolConfig
}

type PoolConfig struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}

type SessionConfig struct {
	Secret   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite string
}

type AuthConfig struct {
	DefaultAdmin DefaultAdminConfig
}

type DefaultAdminConfig struct {
	Username string
	Password string
}

type PprofConfig struct {
	Port int
}

var cfg *Config

func loadYAML(configPath string) *Config {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[Config] YAML file not found at %s, using environment variables and defaults", configPath)
			return nil
		}
		log.Printf("[Config] Error reading YAML file: %v", err)
		return nil
	}

	var yamlCfg Config
	if err := yaml.Unmarshal(data, &yamlCfg); err != nil {
		log.Printf("[Config] Error parsing YAML file: %v", err)
		return nil
	}

	log.Printf("[Config] Successfully loaded configuration from %s", configPath)
	return &yamlCfg
}

func Load() *Config {
	configPath := "config.yaml"

	yamlCfg := loadYAML(configPath)

	if yamlCfg == nil {
		yamlCfg = &Config{}
	}

	secret := getEnv("AG_SERVER_SESSION_SECRET", yamlCfg.Server.Session.Secret)
	if secret == "" {
		secret = generateSecret()
		log.Printf("[Config] Generated random session secret")
	}

	trustedProxies := getStringSlice("AG_SERVER_TRUSTED_PROXIES", yamlCfg.Server.TrustedProxies)
	if len(trustedProxies) == 0 {
		trustedProxies = []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"}
	}

	cfg = &Config{
		Debug: DebugConfig{
			Gin:      getBool("AG_DEBUG_GIN", yamlCfg.Debug.Gin),
			Gorm:     getBool("AG_DEBUG_GORM", yamlCfg.Debug.Gorm),
			Provider: getBool("AG_DEBUG_PROVIDER", yamlCfg.Debug.Provider),
			MCP:      getBool("AG_DEBUG_MCP", yamlCfg.Debug.MCP),
		},
		Server: ServerConfig{
			Port:           getInt("AG_SERVER_PORT", yamlCfg.Server.Port),
			TrustedProxies: trustedProxies,
			Session: SessionConfig{
				Secret:   secret,
				MaxAge:   getInt("AG_SERVER_SESSION_MAX_AGE", yamlCfg.Server.Session.MaxAge),
				Secure:   getBool("AG_SERVER_SESSION_SECURE", yamlCfg.Server.Session.Secure),
				HttpOnly: getBool("AG_SERVER_SESSION_HTTP_ONLY", yamlCfg.Server.Session.HttpOnly),
				SameSite: getEnv("AG_SERVER_SESSION_SAME_SITE", yamlCfg.Server.Session.SameSite),
			},
		},
		Database: DatabaseConfig{
			Type:     getEnv("AG_DATABASE_TYPE", yamlCfg.Database.Type),
			Path:     getEnv("AG_DATABASE_PATH", yamlCfg.Database.Path),
			Host:     getEnv("AG_DATABASE_HOST", yamlCfg.Database.Host),
			Port:     getInt("AG_DATABASE_PORT", yamlCfg.Database.Port),
			Username: getEnv("AG_DATABASE_USERNAME", yamlCfg.Database.Username),
			Password: getEnv("AG_DATABASE_PASSWORD", yamlCfg.Database.Password),
			DBName:   getEnv("AG_DATABASE_DBNAME", yamlCfg.Database.DBName),
			Pool: PoolConfig{
				MaxOpen:     getInt("AG_DATABASE_POOL_MAX_OPEN", yamlCfg.Database.Pool.MaxOpen),
				MaxIdle:     getInt("AG_DATABASE_POOL_MAX_IDLE", yamlCfg.Database.Pool.MaxIdle),
				MaxLifetime: getDuration("AG_DATABASE_POOL_MAX_LIFETIME", yamlCfg.Database.Pool.MaxLifetime),
				MaxIdleTime: getDuration("AG_DATABASE_POOL_MAX_IDLE_TIME", yamlCfg.Database.Pool.MaxIdleTime),
			},
		},
		Auth: AuthConfig{
			DefaultAdmin: DefaultAdminConfig{
				Username: getEnv("AG_ADMIN_USERNAME", yamlCfg.Auth.DefaultAdmin.Username),
				Password: getEnv("AG_ADMIN_PASSWORD", yamlCfg.Auth.DefaultAdmin.Password),
			},
		},
		Pprof: PprofConfig{
			Port: getInt("AG_PPROF_PORT", yamlCfg.Pprof.Port),
		},
	}

	applyDefaults()
	logConfig()

	return cfg
}

func applyDefaults() {
	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite"
		cfg.Database.Path = "data.db"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 18080
	}
	if cfg.Server.Session.MaxAge == 0 {
		cfg.Server.Session.MaxAge = 86400
	}
	if cfg.Server.Session.SameSite == "" {
		cfg.Server.Session.SameSite = "lax"
	}
	if cfg.Auth.DefaultAdmin.Username == "" {
		cfg.Auth.DefaultAdmin.Username = "admin"
	}
	if cfg.Auth.DefaultAdmin.Password == "" {
		cfg.Auth.DefaultAdmin.Password = "admin"
	}
	if cfg.Pprof.Port == 0 {
		cfg.Pprof.Port = 6060
	}

	poolDefaults := getPoolDefaults(cfg.Database.Type)
	if cfg.Database.Pool.MaxOpen == 0 {
		cfg.Database.Pool.MaxOpen = poolDefaults.MaxOpen
	}
	if cfg.Database.Pool.MaxIdle == 0 {
		cfg.Database.Pool.MaxIdle = poolDefaults.MaxIdle
	}
	if cfg.Database.Pool.MaxLifetime == 0 {
		cfg.Database.Pool.MaxLifetime = poolDefaults.MaxLifetime
	}
	if cfg.Database.Pool.MaxIdleTime == 0 {
		cfg.Database.Pool.MaxIdleTime = poolDefaults.MaxIdleTime
	}

	if cfg.Database.Type == "sqlite" {
		if cfg.Database.Pool.MaxOpen > 1 {
			log.Printf("[Config] Warning: SQLite connection pool constrained to MaxOpen=1")
			cfg.Database.Pool.MaxOpen = 1
		}
		if cfg.Database.Pool.MaxIdle > 1 {
			log.Printf("[Config] Warning: SQLite connection pool constrained to MaxIdle=1")
			cfg.Database.Pool.MaxIdle = 1
		}
	}
}

func getPoolDefaults(dbType string) PoolConfig {
	if dbType == "sqlite" {
		return PoolConfig{
			MaxOpen:     1,
			MaxIdle:     1,
			MaxLifetime: time.Hour,
			MaxIdleTime: 5 * time.Minute,
		}
	}
	return PoolConfig{
		MaxOpen:     20,
		MaxIdle:     5,
		MaxLifetime: time.Hour,
		MaxIdleTime: 5 * time.Minute,
	}
}

func Get() *Config {
	return cfg
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func getBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
		if i, err := strconv.Atoi(val); err == nil {
			return time.Duration(i) * time.Second
		}
	}
	return defaultValue
}

func getStringSlice(key string, defaultValue []string) []string {
	if val := os.Getenv(key); val != "" {
		result := []string{}
		for _, item := range splitString(val, ",") {
			if trimmed := trimSpace(item); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultValue
}

func splitString(s, sep string) []string {
	return strings.Split(s, sep)
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func logConfig() {
	log.Println("[Config] Configuration loaded:")
	log.Printf("  Debug Gin: %v", cfg.Debug.Gin)
	log.Printf("  Debug Gorm: %v", cfg.Debug.Gorm)
	log.Printf("  Debug Provider: %v", cfg.Debug.Provider)
	log.Printf("  Debug MCP: %v", cfg.Debug.MCP)
	log.Printf("  Server Port: %d", cfg.Server.Port)
	log.Printf("  Trusted Proxies: %v", cfg.Server.TrustedProxies)
	log.Printf("  Database Type: %s", cfg.Database.Type)
	if cfg.Database.Type == "sqlite" {
		log.Printf("  Database Path: %s", cfg.Database.Path)
	} else {
		log.Printf("  Database Host: %s", cfg.Database.Host)
		log.Printf("  Database Port: %d", cfg.Database.Port)
		log.Printf("  Database Username: %s", cfg.Database.Username)
		log.Printf("  Database Password: %s", maskPassword(cfg.Database.Password))
		log.Printf("  Database Name: %s", cfg.Database.DBName)
	}
	log.Printf("  Database Pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v, MaxIdleTime=%v",
		cfg.Database.Pool.MaxOpen, cfg.Database.Pool.MaxIdle, cfg.Database.Pool.MaxLifetime, cfg.Database.Pool.MaxIdleTime)
	log.Printf("  Session MaxAge: %d", cfg.Server.Session.MaxAge)
	log.Printf("  Session Secure: %v", cfg.Server.Session.Secure)
	log.Printf("  Session HttpOnly: %v", cfg.Server.Session.HttpOnly)
	log.Printf("  Session SameSite: %s", cfg.Server.Session.SameSite)
	log.Printf("  Admin Username: %s", cfg.Auth.DefaultAdmin.Username)
	log.Printf("  Admin Password: %s", maskPassword(cfg.Auth.DefaultAdmin.Password))
	log.Printf("  Pprof Port: %d", cfg.Pprof.Port)
}

func maskPassword(p string) string {
	if len(p) <= 2 {
		return "****"
	}
	return p[:1] + "****" + p[len(p)-1:]
}

func (c *Config) DSN() string {
	switch c.Database.Type {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
			c.Database.Host, c.Database.Port, c.Database.Username, c.Database.Password, c.Database.DBName)
	case "sqlite":
		return fmt.Sprintf("%s?_loc=auto", c.Database.Path)
	default:
		return fmt.Sprintf("%s?_loc=auto", c.Database.Path)
	}
}
