package devops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// GenerateDocker creates a production-grade multi-stage Dockerfile and .dockerignore
func GenerateDocker(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	var dockerfile string
	var dockerignore string

	switch cfg.Framework.ID {
	case "laravel":
		dockerfile = `FROM php:8.3-fpm-alpine

# Install system dependencies
RUN apk add --no-cache \
    git \
    curl \
    libpng-dev \
    libxml2-dev \
    zip \
    unzip \
    postgresql-dev \
    oniguruma-dev \
    bash

# Install PHP extensions
RUN docker-php-ext-install pdo pdo_pgsql pdo_mysql mbstring exif pcntl bcmath gd opcache

# Get latest Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# Set working directory
WORKDIR /var/www/html

# Copy application files
COPY . .

# Copy and set entrypoint script
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Pre-install Composer dependencies during build
RUN composer install --no-dev --no-interaction --prefer-dist --optimize-autoloader || true

# Adjust permissions
RUN chown -R www-data:www-data /var/www/html/storage /var/www/html/bootstrap/cache

EXPOSE 9000
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["php-fpm"]
`
		dockerignore = `/vendor
/node_modules
/.git
.env
.env.backup
.phpunit.result.cache
`

	case "fastapi":
		dockerfile = `FROM python:3.12-slim

# Prevent Python from writing .pyc and buffer stdout
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

EXPOSE 8000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
`
		dockerignore = `__pycache__
*.pyc
*.pyo
*.pyd
.Python
.venv
env/
venv/
.git
.gitignore
.env
`

	case "next":
		dockerfile = `FROM node:20-alpine AS base

# 1. Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./
RUN npm install

# 2. Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# 3. Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
`
		dockerignore = `node_modules
.next
.git
.gitignore
.env
.env*.local
`

	case "fiber":
		dockerfile = `FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server cmd/api/main.go

# Production image
FROM alpine:3.20

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]
`
		dockerignore = `bin/
*.out
*.exe
.git
.gitignore
.env
`

	default:
		dockerfile = fmt.Sprintf(`FROM alpine:latest
WORKDIR /app
COPY . .
EXPOSE %d
CMD ["sh"]
`, cfg.Framework.DefaultPort)
		dockerignore = `.git
.env
node_modules/
vendor/
`
	}

	if err := os.WriteFile(filepath.Join(base, "Dockerfile"), []byte(dockerfile), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, ".dockerignore"), []byte(dockerignore), 0644); err != nil {
		return err
	}

	return nil
}
