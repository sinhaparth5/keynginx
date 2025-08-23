package nginx

import (
	"fmt"
	"github.com/sinhaparth5/keynginx/internal/config"
)

type SecurityProfile struct {
	Name    string
	Headers map[string]string
	Level   string
}

func GetSecurityHeaders(cfg *config.SecurityConfig) map[string]string {
	headers := make(map[string]string)

	headers["X-Server-Created-By"] = "keynginx"

	switch cfg.Level {
	case "strict":
		headers["X-Frame-Options"] = "DENY"
		headers["X-XSS-Protection"] = "1; mode=block"
		headers["X-Content-Type-Options"] = "nosniff"
		headers["Referrer-Policy"] = "no-referrer"
		headers["X-Download-Options"] = "noopen"
		headers["X-Permitted-Cross-Domain-Policies"] = "none"
		if cfg.EnableHSTS {
			headers["Strict-Transport-Security"] = "max-age=63072000; includeSubDomains; preload"
		}
		if cfg.EnableCSP {
			headers["Content-Security-Policy"] = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self';"
		}

	case "balanced":
		headers["X-Frame-Options"] = "SAMEORIGIN"
		headers["X-XSS-Protection"] = "1; mode=block"
		headers["X-Content-Type-Options"] = "nosniff"
		headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
		headers["X-Download-Options"] = "noopen"
		if cfg.EnableHSTS {
			headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
		}
		if cfg.EnableCSP && cfg.CSPPolicy != "" {
			headers["Content-Security-Policy"] = cfg.CSPPolicy
		}

	case "permissive":
		headers["X-Frame-Options"] = "SAMEORIGIN"
		headers["X-Content-Type-Options"] = "nosniff"
		headers["Referrer-Policy"] = "origin-when-cross-origin"
	}

	for key, value := range cfg.CustomHeaders {
		headers[key] = value
	}

	return headers
}

func GetRateLimitConfig(cfg *config.RateLimitConfig) string {
	if !cfg.Enabled {
		return ""
	}

	return fmt.Sprintf(`    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=keynginx:10m rate=%dr/m;
    limit_req zone=keynginx burst=%d nodelay;`, cfg.RequestsPerMinute, cfg.BurstSize)
}
