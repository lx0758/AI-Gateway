package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Debug    DebugConfig
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Auth     AuthConfig
}

type DebugConfig struct {
	Enabled bool
}

type ServerConfig struct {
	Port           int
	Mode           string
	TrustedProxies []string
}

type DatabaseConfig struct {
	Type     string
	Path     string
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
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

	secret := getEnv("AG_SESSION_SECRET", yamlCfg.Session.Secret)
	if secret == "" {
		secret = generateSecret()
		log.Printf("[Config] Generated random session secret")
	}

	trustedProxies := getStringSlice("AG_TRUSTED_PROXIES", yamlCfg.Server.TrustedProxies)
	if len(trustedProxies) == 0 {
		trustedProxies = []string{"10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"}
	}

	cfg = &Config{
		Debug: DebugConfig{
			Enabled: getBool("AG_DEBUG_ENABLED", yamlCfg.Debug.Enabled),
		},
		Server: ServerConfig{
			Port:           getInt("AG_SERVER_PORT", yamlCfg.Server.Port),
			Mode:           getEnv("AG_SERVER_MODE", yamlCfg.Server.Mode),
			TrustedProxies: trustedProxies,
		},
		Database: DatabaseConfig{
			Type:     getEnv("AG_DATABASE_TYPE", yamlCfg.Database.Type),
			Path:     getEnv("AG_DATABASE_PATH", yamlCfg.Database.Path),
			Host:     getEnv("AG_DATABASE_HOST", yamlCfg.Database.Host),
			Port:     getInt("AG_DATABASE_PORT", yamlCfg.Database.Port),
			Username: getEnv("AG_DATABASE_USERNAME", yamlCfg.Database.Username),
			Password: getEnv("AG_DATABASE_PASSWORD", yamlCfg.Database.Password),
			DBName:   getEnv("AG_DATABASE_DBNAME", yamlCfg.Database.DBName),
		},
		Session: SessionConfig{
			Secret:   secret,
			MaxAge:   getInt("AG_SESSION_MAX_AGE", yamlCfg.Session.MaxAge),
			Secure:   getBool("AG_SESSION_SECURE", yamlCfg.Session.Secure),
			HttpOnly: getBool("AG_SESSION_HTTP_ONLY", yamlCfg.Session.HttpOnly),
			SameSite: getEnv("AG_SESSION_SAME_SITE", yamlCfg.Session.SameSite),
		},
		Auth: AuthConfig{
			DefaultAdmin: DefaultAdminConfig{
				Username: getEnv("AG_ADMIN_USERNAME", yamlCfg.Auth.DefaultAdmin.Username),
				Password: getEnv("AG_ADMIN_PASSWORD", yamlCfg.Auth.DefaultAdmin.Password),
			},
		},
	}

	if cfg.Database.Type == "" {
		cfg.Database.Type = "sqlite"
		cfg.Database.Path = "data.db"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 18080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Session.MaxAge == 0 {
		cfg.Session.MaxAge = 86400
	}
	if cfg.Session.SameSite == "" {
		cfg.Session.SameSite = "lax"
	}
	if cfg.Auth.DefaultAdmin.Username == "" {
		cfg.Auth.DefaultAdmin.Username = "admin"
	}
	if cfg.Auth.DefaultAdmin.Password == "" {
		cfg.Auth.DefaultAdmin.Password = "admin"
	}

	logConfig()

	return cfg
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
	log.Printf("  Debug Enabled: %v", cfg.Debug.Enabled)
	log.Printf("  Server Port: %d", cfg.Server.Port)
	log.Printf("  Server Mode: %s", cfg.Server.Mode)
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
	log.Printf("  Session MaxAge: %d", cfg.Session.MaxAge)
	log.Printf("  Session Secure: %v", cfg.Session.Secure)
	log.Printf("  Session HttpOnly: %v", cfg.Session.HttpOnly)
	log.Printf("  Session SameSite: %s", cfg.Session.SameSite)
	log.Printf("  Admin Username: %s", cfg.Auth.DefaultAdmin.Username)
	log.Printf("  Admin Password: %s", maskPassword(cfg.Auth.DefaultAdmin.Password))
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
