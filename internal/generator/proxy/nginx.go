package proxy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// GenerateNginx creates standard Nginx reverse proxy configuration tailored to the runtime
func GenerateNginx(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath
	nginxDir := filepath.Join(base, "docker/nginx")
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		return err
	}

	var conf string
	if cfg.Framework.ID == "laravel" {
		conf = `server {
    listen 80;
    server_name localhost;
    root /var/www/html/public;
    index index.php index.html;

    charset utf-8;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";
    add_header X-XSS-Protection "1; mode=block";

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location = /favicon.ico { access_log off; log_not_found off; }
    location = /robots.txt  { access_log off; log_not_found off; }

    error_page 404 /index.php;

    location ~ \.php$ {
        fastcgi_pass app:9000;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
        fastcgi_hide_header X-Powered-By;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }
}
`
	} else {
		appPort := cfg.Framework.DefaultPort
		if appPort <= 0 {
			appPort = 8080
		}
		conf = fmt.Sprintf(`upstream app_backend {
    server app:%d;
    keepalive 32;
}

server {
    listen 80;
    server_name localhost;

    # Gzip Compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml+rss text/javascript;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://app_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
`, appPort)
	}

	return os.WriteFile(filepath.Join(nginxDir, "default.conf"), []byte(conf), 0644)
}
