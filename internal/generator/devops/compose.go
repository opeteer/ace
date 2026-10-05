package devops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// GenerateCompose dynamically builds docker-compose.yml based on framework, db, and proxy options
func GenerateCompose(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	var sb strings.Builder
	sb.WriteString("services:\n")

	// 1. App Service
	sb.WriteString("  app:\n")
	sb.WriteString("    build:\n")
	sb.WriteString("      context: .\n")
	sb.WriteString("      dockerfile: Dockerfile\n")
	sb.WriteString("    restart: unless-stopped\n")
	sb.WriteString("    env_file:\n")
	sb.WriteString("      - .env\n")

	// App volumes & ports
	switch cfg.Framework.ID {
	case "laravel":
		sb.WriteString("    volumes:\n")
		sb.WriteString("      - ./:/var/www/html\n")
		sb.WriteString("      - app_vendor:/var/www/html/vendor\n")
	case "fastapi", "django":
		sb.WriteString("    volumes:\n")
		sb.WriteString("      - ./:/app\n")
	case "fiber", "gin":
		sb.WriteString("    volumes:\n")
		sb.WriteString("      - ./:/app\n")
	}

	// Port exposure: If no reverse proxy, expose app port directly
	if cfg.Proxy == "" || cfg.Proxy == config.ProxyNone {
		appPort := cfg.Framework.DefaultPort
		if appPort > 0 {
			sb.WriteString("    ports:\n")
			sb.WriteString(fmt.Sprintf("      - \"%d:%d\"\n", appPort, appPort))
		}
	}

	if cfg.Database.ID != "" && cfg.Database.ID != "sqlite" {
		sb.WriteString("    depends_on:\n")
		sb.WriteString("      db:\n")
		sb.WriteString("        condition: service_healthy\n")
	}

	// 2. Database Service
	if cfg.Database.ID != "" && cfg.Database.ID != "sqlite" {
		sb.WriteString("\n  db:\n")
		sb.WriteString(fmt.Sprintf("    image: %s\n", cfg.Database.DockerImage))
		sb.WriteString("    restart: unless-stopped\n")

		switch cfg.Database.ID {
		case "postgres":
			sb.WriteString("    environment:\n")
			sb.WriteString(fmt.Sprintf("      POSTGRES_DB: %s\n", cfg.DBName()))
			sb.WriteString("      POSTGRES_USER: ace_user\n")
			sb.WriteString("      POSTGRES_PASSWORD: secret\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"5432:5432\"\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - db_data:/var/lib/postgresql/data\n")
			sb.WriteString("    healthcheck:\n")
			sb.WriteString(fmt.Sprintf("      test: [\"CMD-SHELL\", \"pg_isready -U ace_user -d %s\"]\n", cfg.DBName()))
			sb.WriteString("      interval: 5s\n")
			sb.WriteString("      timeout: 5s\n")
			sb.WriteString("      retries: 5\n")

		case "mysql", "mariadb":
			sb.WriteString("    environment:\n")
			sb.WriteString(fmt.Sprintf("      MYSQL_DATABASE: %s\n", cfg.DBName()))
			sb.WriteString("      MYSQL_USER: ace_user\n")
			sb.WriteString("      MYSQL_PASSWORD: secret\n")
			sb.WriteString("      MYSQL_ROOT_PASSWORD: rootsecret\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"3306:3306\"\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - db_data:/var/lib/mysql\n")
			sb.WriteString("    healthcheck:\n")
			sb.WriteString("      test: [\"CMD\", \"mysqladmin\", \"ping\", \"-h\", \"localhost\", \"-u\", \"ace_user\", \"-psecret\"]\n")
			sb.WriteString("      interval: 5s\n")
			sb.WriteString("      timeout: 5s\n")
			sb.WriteString("      retries: 5\n")

		case "mongo":
			sb.WriteString("    environment:\n")
			sb.WriteString("      MONGO_INITDB_ROOT_USERNAME: ace_user\n")
			sb.WriteString("      MONGO_INITDB_ROOT_PASSWORD: secret\n")
			sb.WriteString(fmt.Sprintf("      MONGO_INITDB_DATABASE: %s\n", cfg.DBName()))
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"27017:27017\"\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - db_data:/data/db\n")
			sb.WriteString("    healthcheck:\n")
			sb.WriteString("      test: [\"CMD\", \"mongosh\", \"--eval\", \"db.adminCommand('ping')\"]\n")
			sb.WriteString("      interval: 5s\n")
			sb.WriteString("      timeout: 5s\n")
			sb.WriteString("      retries: 5\n")

		case "redis":
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"6379:6379\"\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - db_data:/data\n")
			sb.WriteString("    healthcheck:\n")
			sb.WriteString("      test: [\"CMD\", \"redis-cli\", \"ping\"]\n")
			sb.WriteString("      interval: 5s\n")
			sb.WriteString("      timeout: 5s\n")
			sb.WriteString("      retries: 5\n")
		}
	}

	// 3. Reverse Proxy (Nginx)
	if cfg.Proxy == config.ProxyNginx {
		sb.WriteString("\n  proxy:\n")
		sb.WriteString("    image: nginx:alpine\n")
		sb.WriteString("    restart: unless-stopped\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"80:80\"\n")
		sb.WriteString("    volumes:\n")
		sb.WriteString("      - ./docker/nginx:/etc/nginx/conf.d:ro\n")
		if cfg.Framework.ID == "laravel" {
			sb.WriteString("      - ./:/var/www/html:ro\n")
		}
		sb.WriteString("    depends_on:\n")
		sb.WriteString("      - app\n")
	}

	// Volumes section
	hasVolumes := false
	if cfg.Database.ID != "" && cfg.Database.ID != "sqlite" {
		hasVolumes = true
	}
	if cfg.Framework.ID == "laravel" {
		hasVolumes = true
	}

	if hasVolumes {
		sb.WriteString("\nvolumes:\n")
		if cfg.Database.ID != "" && cfg.Database.ID != "sqlite" {
			sb.WriteString("  db_data:\n")
		}
		if cfg.Framework.ID == "laravel" {
			sb.WriteString("  app_vendor:\n")
		}
	}

	return os.WriteFile(filepath.Join(base, "docker-compose.yml"), []byte(sb.String()), 0644)
}
