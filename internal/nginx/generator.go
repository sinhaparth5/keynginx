package nginx

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/sinhaparth5/keynginx/internal/config"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateConfig(cfg *config.Config) (string, error) {
	tmpl, err := template.New("nginx").Parse(nginxTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse nginx template: %w", err)
	}

	data := struct {
		*config.Config
		SecurityHeaders map[string]string
		RateLimitConfig string
		Timestamp       string
	}{
		Config:          cfg,
		SecurityHeaders: GetSecurityHeaders(&cfg.Security),
		RateLimitConfig: GetRateLimitConfig(&cfg.Security.RateLimit),
		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute nginx template: %w", err)
	}

	return buf.String(), nil
}

func (g *Generator) GenerateDockerCompose(cfg *config.Config) (string, error) {
	tmpl, err := template.New("docker-compose").Parse(dockerComposeTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse docker-compose template: %w", err)
	}

	data := struct {
		*config.Config
		Timestamp string
	}{
		Config:    cfg,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute docker-compose template: %w", err)
	}

	return buf.String(), nil
}

const nginxTemplate = `# KeyNginx Generated Configuration
# Domain: {{.Project.Domain}}
# Security Level: {{.Security.Level}}
# Generated: {{.Timestamp}}

events {
    worker_connections 1024;
    multi_accept on;
    use epoll;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    # Logging Configuration
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                   '$status $body_bytes_sent "$http_referer" '
                   '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;

    # Basic Settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 16M;

    # Hide server tokens
    server_tokens off;

    # Gzip Compression
    gzip on;
    gzip_vary on;
    gzip_comp_level 6;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/json
        application/xml
        application/rss+xml
        application/atom+xml
        image/svg+xml;

{{.RateLimitConfig}}

    # HTTP to HTTPS redirect
    server {
        listen {{.Nginx.HTTPPort}};
        server_name {{.Nginx.ServerName}};
        return 301 https://$server_name$request_uri;
    }

    # HTTPS server
    server {
        listen {{.Nginx.HTTPSPort}} ssl http2;
        server_name {{.Nginx.ServerName}};

        # SSL Configuration
        ssl_certificate /etc/nginx/ssl/certificate.crt;
        ssl_certificate_key /etc/nginx/ssl/private.key;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;
        ssl_session_cache shared:SSL:10m;
        ssl_session_timeout 10m;

        # Security Headers{{range $key, $value := .SecurityHeaders}}
        add_header {{$key}} "{{$value}}" always;{{end}}

        # Custom Headers{{range $key, $value := .Nginx.CustomHeaders}}
        add_header {{$key}} "{{$value}}" always;{{end}}

{{if .Nginx.Services}}{{range .Nginx.Services}}
        # Service: {{.Name}}
        location {{.Path}} {
            proxy_pass {{.ProxyPass}};
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-Host $server_name;

            # Remove server identification
            proxy_hide_header X-Powered-By;
            proxy_hide_header Server;
        }
{{end}}{{else}}
        # Default location
        location / {
            root /usr/share/nginx/html;
            index index.html index.htm;
        }
{{end}}
        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }

        # Security.txt endpoint
        location /.well-known/security.txt {
            return 200 "# KeyNginx Generated Security Policy\nContact: mailto:admin@{{.Nginx.ServerName}}\n";
            add_header Content-Type text/plain;
        }
    }
}
`

const dockerComposeTemplate = `# KeyNginx Generated Docker Compose
# Domain: {{.Project.Domain}}
# Generated: {{.Timestamp}}
version: '{{.Docker.ComposeVersion}}'

services:
  nginx:
    image: {{.Docker.NginxImage}}
    container_name: keynginx-{{.Project.Domain}}
    ports:
      - "{{.Nginx.HTTPSPort}}:443"
      - "{{.Nginx.HTTPPort}}:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - ./logs:/var/log/nginx
    restart: unless-stopped
    networks:
      - {{.Docker.NetworkName}}

{{if .Nginx.Services}}{{range .Nginx.Services}}
  # {{.Name}}:
  #   build: ./{{.Name}}
  #   ports:
  #     - "{{.Port}}:{{.Port}}"
  #   networks:
  #     - {{$.Docker.NetworkName}}
  #   # Uncomment and configure as needed

{{end}}{{end}}
networks:
  {{.Docker.NetworkName}}:
    driver: bridge

volumes:
  nginx-logs:
`
