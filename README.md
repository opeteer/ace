# Ace CLI (Universal Scaffolding Engine)

> **One tool to scaffold, containerize, and wire everything.**

Ace is a fast, polyglot CLI tool for automated project scaffolding across web, mobile, and backend frameworks. It removes boilerplate fatigue by instantly generating production-ready project structures with pre-configured databases, Docker & Docker Compose orchestration, Nginx reverse proxies, and CI/CD pipelines.

---

## Features

- **19+ Frameworks**: Web Fullstack (Next.js, Laravel, Nuxt, SvelteKit, Astro, Django, Rails), Backend APIs (FastAPI, Go Fiber/Gin, Express, NestJS, Axum, Spring Boot, ASP.NET Core), Frontend SPAs (React, Vue), and Cross-Platform Mobile (Flutter, React Native/Expo).
- **10+ Database Engines**: PostgreSQL, MySQL, MariaDB, SQLite, MongoDB, Redis/Valkey, ClickHouse, Meilisearch, Elasticsearch, Neo4j.
- **Smart ORM / Driver Injection**: Automatically configures the idiomatic data layer (Prisma/Drizzle for TS, Eloquent for Laravel, SQLAlchemy for FastAPI, GORM for Go Fiber).
- **Containerization on Demand**: Multi-stage, language-specific `Dockerfile`s and orchestrated `docker-compose.yml` with healthchecks, persistent volumes, and networks.
- **Reverse Proxy**: Tailored Nginx configurations for FastCGI (PHP-FPM) and HTTP reverse proxy with WebSocket upgrade support.
- **CI/CD Built-in**: Plug-and-play GitHub Actions workflows for matrix tests and container builds.
- **Interactive Wizard or Fast One-Liners**: Rich interactive terminal TUI wizard or fast flag-driven CLI commands.

---

## Installation & Build

Build the single static binary:
```bash
go build -o bin/ace main.go
```

Verify installation:
```bash
ace --help
ace list
ace showtime
```

---

## Usage Examples

### 0. Showtime (Signature Spade Logo)
```bash
ace showtime
# or plain monochrome:
ace showtime --mono
```

### 1. Interactive Wizard
Run without flags to open the interactive selection wizard:
```bash
ace new
# or
ace create my-app
```

### 2. Laravel + PostgreSQL + Docker + Nginx + GitHub Actions
```bash
./bin/ace new my-laravel-app \
  -f laravel \
  -d postgres \
  --docker \
  --nginx \
  --ci github
```

### 3. FastAPI + PostgreSQL + Docker
```bash
./bin/ace new my-api-service \
  -f fastapi \
  -d postgres \
  --docker \
  --ci github
```

### 4. Next.js + PostgreSQL (Prisma)
```bash
./bin/ace new my-web-app \
  -f next \
  -d postgres
```

### 5. Go Fiber Microservice + PostgreSQL + Docker
```bash
./bin/ace new my-fiber-service \
  -f fiber \
  -d postgres \
  --docker
```

---

## CLI Flags Reference

| Flag | Short | Description | Default |
| :--- | :--- | :--- | :--- |
| `--framework` | `-f` | Framework ID (e.g. `laravel`, `fastapi`, `next`, `fiber`) | Interactive |
| `--db` | `-d` | Database engine (`postgres`, `mysql`, `mongo`, `sqlite`, `redis`) | None |
| `--docker` | | Generate Dockerfile & docker-compose.yml | `false` |
| `--nginx` | | Shorthand to enable Nginx reverse proxy | `false` |
| `--proxy` | | Reverse proxy type (`nginx`, `caddy`, `none`) | `none` |
| `--ci` | | CI/CD workflow (`github`, `gitlab`, `none`) | `none` |
| `--protocol` | | API protocol (`rest`, `graphql`, `grpc`, `websocket`) | `rest` |
| `--no-git` | | Skip git repository initialization | `false` (git init enabled) |

---

## Testing

Run the test suite:
```bash
go test -v ./...
```
