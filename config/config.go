package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)
	bindEnvVars(v)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

func bindEnvVars(v *viper.Viper) {
	v.BindEnv("env", "ENV")

	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.host", "SERVER_HOST")

	v.BindEnv("sqlite.databasename", "SQLITE_DATABASE_NAME")
	v.BindEnv("sqlite.maxconn", "SQLITE_MAX_CONN")
	v.BindEnv("sqlite.maxidle", "SQLITE_MAX_IDLE")
	v.BindEnv("sqlite.maxlifetime", "SQLITE_MAX_LIFETIME")

	v.BindEnv("ratelimit.maxrequests", "RATE_LIMIT_MAX_REQUESTS")
	v.BindEnv("ratelimit.window", "RATE_LIMIT_WINDOW")

	v.BindEnv("cors.allowedorigins", "CORS_ALLOWED_ORIGINS")
	v.BindEnv("cors.allowedmethods", "CORS_ALLOWED_METHODS")
	v.BindEnv("cors.allowedheaders", "CORS_ALLOWED_HEADERS")

	v.BindEnv("url.apibaseurl", "API_BASE_URL")
	v.BindEnv("url.appbaseurl", "APP_BASE_URL")

	v.BindEnv("resend.apikey", "RESEND_API_KEY")
	v.BindEnv("resend.domain", "RESEND_DOMAIN")
	v.BindEnv("resend.timeout", "RESEND_TIMEOUT")

	v.BindEnv("session.secret", "SESSION_SECRET")
	v.BindEnv("session.duration", "SESSION_DURATION")
	v.BindEnv("session.tokensize", "SESSION_TOKEN_SIZE")
	v.BindEnv("session.cookiename", "SESSION_COOKIE_NAME")
	v.BindEnv("session.cookiesecure", "SESSION_COOKIE_SECURE")
	v.BindEnv("session.cookiesamesite", "SESSION_COOKIE_SAME_SITE")

	v.BindEnv("auth.method", "AUTH_METHOD")
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("env", "development")

	v.SetDefault("server.port", 5001)
	v.SetDefault("server.host", "localhost")

	v.SetDefault("sqlite.databasename", "users.db")
	v.SetDefault("sqlite.maxconn", 10)
	v.SetDefault("sqlite.maxidle", 5)
	v.SetDefault("sqlite.maxlifetime", "300s")

	v.SetDefault("ratelimit.maxrequests", 100)
	v.SetDefault("ratelimit.window", "1m")

	v.SetDefault("cors.allowedorigins", []string{"*"})
	v.SetDefault("cors.allowedmethods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowedheaders", []string{"Content-Type", "Authorization"})

	v.SetDefault("url.apibaseurl", "http://localhost:5001")
	v.SetDefault("url.appbaseurl", "http://localhost:5173")

	v.SetDefault("resend.timeout", "10s")

	v.SetDefault("session.secret", "cjQ6A2CJ2V5g2StB6DPYA3rxfvOlKm3m")
	v.SetDefault("session.duration", "168h")
	v.SetDefault("session.tokensize", 32)
	v.SetDefault("session.cookiename", "base-project:session")
	v.SetDefault("session.cookiesecure", false)
	v.SetDefault("session.cookiesamesite", "strict")

	v.SetDefault("auth.method", "password")
}
