package middleware

import (
	"time"

	"github.com/g-villarinho/voxel-api/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// RateLimiter returns rate limiting middleware configured from the application config.
func RateLimiter(cfg *config.Config) echo.MiddlewareFunc {
	rateLimit := rate.Limit(cfg.RateLimit.MaxRequests)
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rateLimit,
			Burst:     cfg.RateLimit.MaxRequests,
			ExpiresIn: cfg.RateLimit.Window,
		},
	))
}

// AuthRateLimiter returns a stricter rate limiter for auth routes that send emails.
// Limits to 5 requests per IP per 15 minutes to prevent email spam attacks.
func AuthRateLimiter() echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(5.0 / 900.0), // 5 requests per 900 seconds (15 min)
			Burst:     5,
			ExpiresIn: 15 * time.Minute,
		},
	))
}
