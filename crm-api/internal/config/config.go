package config

import (
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig
	DB      DBConfig
	Redis   RedisConfig
	JWT     JWTConfig
	Storage StorageConfig
	SMTP    SMTPConfig
	Log     LogConfig
}

type AppConfig struct {
	Env  string
	Port string
	Host string
	// AllowedOrigins lists other web origins that may call the API from a
	// browser. Empty means same-origin only, which is how nginx serves the app.
	AllowedOrigins []string
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	RefreshSecret  string
	AccessExpiry   time.Duration
	RefreshExpiry  time.Duration
}

type StorageConfig struct {
	Root          string
	MaxUploadSize int64
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type LogConfig struct {
	Level string
	File  string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = viper.ReadInConfig()

	setDefaults()

	cfg := &Config{
		App: AppConfig{
			Env:            viper.GetString("APP_ENV"),
			Port:           viper.GetString("APP_PORT"),
			Host:           viper.GetString("APP_HOST"),
			AllowedOrigins: splitList(viper.GetString("CORS_ALLOWED_ORIGINS")),
		},
		DB: DBConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			Name:     viper.GetString("DB_NAME"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("REDIS_ADDR"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			PrivateKeyPath: viper.GetString("JWT_PRIVATE_KEY_PATH"),
			PublicKeyPath:  viper.GetString("JWT_PUBLIC_KEY_PATH"),
			RefreshSecret:  viper.GetString("JWT_REFRESH_SECRET"),
			AccessExpiry:   viper.GetDuration("JWT_ACCESS_EXPIRY"),
			RefreshExpiry:  viper.GetDuration("JWT_REFRESH_EXPIRY"),
		},
		Storage: StorageConfig{
			Root:          viper.GetString("STORAGE_ROOT"),
			MaxUploadSize: viper.GetInt64("MAX_UPLOAD_SIZE"),
		},
		SMTP: SMTPConfig{
			Host:     viper.GetString("SMTP_HOST"),
			Port:     viper.GetInt("SMTP_PORT"),
			User:     viper.GetString("SMTP_USER"),
			Password: viper.GetString("SMTP_PASS"),
			From:     viper.GetString("SMTP_FROM"),
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
			File:  viper.GetString("LOG_FILE"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func splitList(value string) []string {
	items := []string{}
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

const defaultRefreshSecret = "change-me-in-production"

// validate refuses to run a production server on development fallbacks, so a
// missing .env value stops the deploy instead of silently weakening it.
func (c *Config) validate() error {
	if c.App.Env != "production" {
		return nil
	}
	var problems []string
	if strings.Contains(c.JWT.RefreshSecret, "change-me") || len(c.JWT.RefreshSecret) < 32 {
		problems = append(problems, "JWT_REFRESH_SECRET must be set to a random value of at least 32 characters")
	}
	if c.DB.Password == "" || c.DB.Password == "devpassword" || strings.Contains(c.DB.Password, "change-me") {
		problems = append(problems, "DB_PASSWORD must be set to a non-default value")
	}
	for _, origin := range c.App.AllowedOrigins {
		if origin == "*" {
			problems = append(problems, "CORS_ALLOWED_ORIGINS must list specific origins, not *")
		}
	}
	if len(problems) > 0 {
		return errors.New("unsafe production configuration: " + strings.Join(problems, "; "))
	}
	return nil
}

func (d *DBConfig) DSN() string {
	connection := url.URL{Scheme: "postgres", User: url.UserPassword(d.User, d.Password), Host: net.JoinHostPort(d.Host, d.Port), Path: "/" + d.Name}
	query := connection.Query()
	query.Set("sslmode", d.SSLMode)
	query.Set("TimeZone", "UTC")
	connection.RawQuery = query.Encode()
	return connection.String()
}

func setDefaults() {
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_HOST", "0.0.0.0")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_NAME", "g4s_crm")
	viper.SetDefault("DB_USER", "g4s_app")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_SSLMODE", "disable")

	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_PRIVATE_KEY_PATH", "./keys/private.pem")
	viper.SetDefault("JWT_PUBLIC_KEY_PATH", "./keys/public.pem")
	viper.SetDefault("JWT_REFRESH_SECRET", defaultRefreshSecret)
	viper.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRY", "168h")

	viper.SetDefault("STORAGE_ROOT", "./storage")
	viper.SetDefault("MAX_UPLOAD_SIZE", 52428800) // 50 MB

	viper.SetDefault("SMTP_PORT", 587)

	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FILE", "")
}
