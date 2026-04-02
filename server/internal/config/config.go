package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Path string
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

func Load() *Config {
	secret := getEnv("AG_SESSION_SECRET", "")
	if secret == "" {
		secret = generateSecret()
		log.Printf("[Config] Generated random session secret")
	}

	cfg = &Config{
		Server: ServerConfig{
			Port: getInt("AG_SERVER_PORT", 18080),
			Mode: getEnv("AG_SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Path: getEnv("AG_DATABASE_PATH", "data.db"),
		},
		Session: SessionConfig{
			Secret:   secret,
			MaxAge:   getInt("AG_SESSION_MAX_AGE", 86400),
			Secure:   getBool("AG_SESSION_SECURE", false),
			HttpOnly: getBool("AG_SESSION_HTTP_ONLY", true),
			SameSite: getEnv("AG_SESSION_SAME_SITE", "lax"),
		},
		Auth: AuthConfig{
			DefaultAdmin: DefaultAdminConfig{
				Username: getEnv("AG_ADMIN_USERNAME", "admin"),
				Password: getEnv("AG_ADMIN_PASSWORD", "admin"),
			},
		},
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

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func logConfig() {
	log.Println("[Config] Configuration loaded:")
	log.Printf("  Server Port: %d", cfg.Server.Port)
	log.Printf("  Server Mode: %s", cfg.Server.Mode)
	log.Printf("  Database Path: %s", cfg.Database.Path)
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
	return fmt.Sprintf("%s?_loc=auto", c.Database.Path)
}
