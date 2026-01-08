package config

import "time"

const (
	Development = "development"
	Staging     = "staging"
	Production  = "production"
)

const (
	AuthMethodPassword  = "password"
	AuthMethodMagicLink = "magic_link"
	AuthMethodBoth      = "both"
)

type Config struct {
	Env       string    `mapstructure:"env"`
	Server    Server    `mapstructure:"server"`
	RateLimit RateLimit `mapstructure:"ratelimit"`
	Cors      Cors      `mapstructure:"cors"`
	URL       URL       `mapstructure:"url"`
	Database  Database  `mapstructure:"database"`
	Resend    Resend    `mapstructure:"resend"`
	Session   Session   `mapstructure:"session"`
	Auth      Auth      `mapstructure:"auth"`
}

type Auth struct {
	Method string `mapstructure:"method"`
}

type Server struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type Database struct {
	DSN         string        `mapstructure:"dsn"`
	MaxConn     int32         `mapstructure:"maxconn"`
	MaxIdle     int32         `mapstructure:"maxidle"`
	MaxLifeTime time.Duration `mapstructure:"maxlifetime"`
}

type RateLimit struct {
	MaxRequests int           `mapstructure:"maxrequests"`
	Window      time.Duration `mapstructure:"window"`
}

type Cors struct {
	AllowedOrigins []string `mapstructure:"allowedorigins"`
	AllowedMethods []string `mapstructure:"allowedmethods"`
	AllowedHeaders []string `mapstructure:"allowedheaders"`
}

type URL struct {
	APIBaseURL string `mapstructure:"apibaseurl"`
	APPBaseURL string `mapstructure:"appbaseurl"`
}

type Resend struct {
	APIKey  string        `mapstructure:"apikey"`
	Domain  string        `mapstructure:"domain"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type Session struct {
	Secret         string        `mapstructure:"secret"`
	Duration       time.Duration `mapstructure:"duration"`
	TokenSize      int           `mapstructure:"tokensize"`
	CookieName     string        `mapstructure:"cookiename"`
	CookieSecure   bool          `mapstructure:"cookiesecure"`
	CookieSameSite string        `mapstructure:"cookiesamesite"`
}

func (e *Config) IsDevelopment() bool {
	return e.Env == Development
}

func (e *Config) IsStaging() bool {
	return e.Env == Staging
}

func (e *Config) IsProduction() bool {
	return e.Env == Production
}

func (c *Config) IsPasswordAuthEnabled() bool {
	return c.Auth.Method == AuthMethodPassword || c.Auth.Method == AuthMethodBoth
}

func (c *Config) IsMagicLinkAuthEnabled() bool {
	return c.Auth.Method == AuthMethodMagicLink || c.Auth.Method == AuthMethodBoth
}

func (c *Config) IsPasswordRequired() bool {
	return c.Auth.Method == AuthMethodPassword
}
