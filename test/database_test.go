package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/doctor"
	"github.com/opeteer/ace/internal/generator"
)

func TestScaffoldDjangoMySQLAndRedis(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "django-mysql-redis")

	fw, _ := config.FindFramework("django")
	db, _ := config.FindDatabase("mysql")

	cfg := &config.ProjectConfig{
		Name:        "django-mysql-redis",
		TargetPath:  targetPath,
		Framework:   fw,
		Database:    db,
		Redis:       true,
		Docker:      true,
		NoInstall:   true,
		NoGit:       true,
		Interactive: false,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute failed: %v", err)
	}

	// 1. Verify requirements.txt has both pymysql and django-redis
	reqBytes, err := os.ReadFile(filepath.Join(targetPath, "requirements.txt"))
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}
	reqContent := string(reqBytes)
	if !strings.Contains(reqContent, "pymysql") {
		t.Error("requirements.txt missing pymysql driver for MySQL")
	}
	if !strings.Contains(reqContent, "django-redis") {
		t.Error("requirements.txt missing django-redis driver for Redis")
	}

	// 2. Verify pymysql monkeypatch in package __init__.py
	initBytes, err := os.ReadFile(filepath.Join(targetPath, "django_mysql_redis", "__init__.py"))
	if err != nil {
		t.Fatalf("Failed to read __init__.py: %v", err)
	}
	if !strings.Contains(string(initBytes), "pymysql.install_as_MySQLdb()") {
		t.Error("__init__.py missing pymysql.install_as_MySQLdb()")
	}

	// 3. Verify .env has MySQL and Redis configuration
	envBytes, err := os.ReadFile(filepath.Join(targetPath, ".env"))
	if err != nil {
		t.Fatalf("Failed to read .env: %v", err)
	}
	envContent := string(envBytes)
	if !strings.Contains(envContent, "DB_CONNECTION=mysql") {
		t.Error(".env missing DB_CONNECTION=mysql")
	}
	if !strings.Contains(envContent, "DB_ENGINE=django.db.backends.mysql") {
		t.Error(".env missing DB_ENGINE=django.db.backends.mysql")
	}
	if !strings.Contains(envContent, "REDIS_HOST=redis") {
		t.Error(".env missing REDIS_HOST=redis")
	}
	if !strings.Contains(envContent, "REDIS_PORT=6379") {
		t.Error(".env missing REDIS_PORT=6379")
	}

	// 4. Verify docker-compose.yml has both mysql and redis services
	composeBytes, err := os.ReadFile(filepath.Join(targetPath, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}
	composeContent := string(composeBytes)
	if !strings.Contains(composeContent, "image: mysql:8.4") {
		t.Error("docker-compose.yml missing mysql:8.0 image")
	}
	if !strings.Contains(composeContent, "image: redis:7-alpine") {
		t.Error("docker-compose.yml missing redis:7-alpine image")
	}
	if !strings.Contains(composeContent, "- db") || !strings.Contains(composeContent, "- redis") {
		t.Error("docker-compose.yml app depends_on missing db or redis")
	}

	// 5. Scoped Doctor Check
	report, err := doctor.DiagnoseProject(targetPath)
	if err != nil {
		t.Fatalf("DiagnoseProject failed: %v", err)
	}

	foundDBConfig := false
	foundMySQLDriver := false
	foundRedisCache := false
	foundRedisDriver := false

	for _, check := range report.Checks {
		if check.Name == "Database Config" && strings.Contains(check.Current, "MySQL") {
			foundDBConfig = true
		}
		if check.Name == "MySQL Driver" {
			foundMySQLDriver = true
		}
		if check.Name == "Redis Cache" {
			foundRedisCache = true
		}
		if check.Name == "Redis Driver" {
			foundRedisDriver = true
		}
	}

	if !foundDBConfig {
		t.Error("Project doctor did not report MySQL Database Config")
	}
	if !foundMySQLDriver {
		t.Error("Project doctor did not detect MySQL Driver")
	}
	if !foundRedisCache {
		t.Error("Project doctor did not report Redis Cache")
	}
	if !foundRedisDriver {
		t.Error("Project doctor did not detect Redis Driver")
	}
}

func TestScaffoldFastAPIMySQLAndRedis(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "fastapi-mysql-redis")

	fw, _ := config.FindFramework("fastapi")
	db, _ := config.FindDatabase("mysql")

	cfg := &config.ProjectConfig{
		Name:        "fastapi-mysql-redis",
		TargetPath:  targetPath,
		Framework:   fw,
		Database:    db,
		Redis:       true,
		Docker:      true,
		NoInstall:   true,
		NoGit:       true,
		Interactive: false,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute failed: %v", err)
	}

	// 1. Verify requirements.txt
	reqBytes, err := os.ReadFile(filepath.Join(targetPath, "requirements.txt"))
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}
	reqContent := string(reqBytes)
	if !strings.Contains(reqContent, "aiomysql") {
		t.Error("requirements.txt missing aiomysql driver for MySQL")
	}
	if !strings.Contains(reqContent, "redis") {
		t.Error("requirements.txt missing redis client")
	}

	// 2. Verify app/core/redis.py was created
	redisPyPath := filepath.Join(targetPath, "app", "core", "redis.py")
	if _, err := os.Stat(redisPyPath); os.IsNotExist(err) {
		t.Errorf("Expected helper '%s' was not generated", redisPyPath)
	}

	// 3. Verify docker-compose.yml
	composeBytes, err := os.ReadFile(filepath.Join(targetPath, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}
	composeContent := string(composeBytes)
	if !strings.Contains(composeContent, "image: mysql:8.4") {
		t.Error("docker-compose.yml missing mysql image")
	}
	if !strings.Contains(composeContent, "image: redis:7-alpine") {
		t.Error("docker-compose.yml missing redis image")
	}

	// 4. Project Doctor
	report, err := doctor.DiagnoseProject(targetPath)
	if err != nil {
		t.Fatalf("DiagnoseProject failed: %v", err)
	}

	foundDBConfig := false
	foundRedisCache := false
	for _, check := range report.Checks {
		if check.Name == "Database Config" && strings.Contains(check.Current, "MySQL") {
			foundDBConfig = true
		}
		if check.Name == "Redis Cache" {
			foundRedisCache = true
		}
	}
	if !foundDBConfig {
		t.Error("Project doctor did not report MySQL Database Config for FastAPI")
	}
	if !foundRedisCache {
		t.Error("Project doctor did not report Redis Cache for FastAPI")
	}
}

func TestScaffoldDjangoSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "django-sqlite")

	fw, _ := config.FindFramework("django")
	db, _ := config.FindDatabase("sqlite")

	cfg := &config.ProjectConfig{
		Name:        "django-sqlite",
		TargetPath:  targetPath,
		Framework:   fw,
		Database:    db,
		Docker:      false,
		NoInstall:   true,
		NoGit:       true,
		Interactive: false,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute failed: %v", err)
	}

	envBytes, err := os.ReadFile(filepath.Join(targetPath, ".env"))
	if err != nil {
		t.Fatalf("Failed to read .env: %v", err)
	}
	if !strings.Contains(string(envBytes), "DB_ENGINE=django.db.backends.sqlite3") {
		t.Error(".env missing DB_ENGINE=django.db.backends.sqlite3")
	}

	report, err := doctor.DiagnoseProject(targetPath)
	if err != nil {
		t.Fatalf("DiagnoseProject failed: %v", err)
	}

	foundSQLite := false
	for _, check := range report.Checks {
		if check.Name == "Database Config" && strings.Contains(check.Current, "SQLite3") {
			foundSQLite = true
		}
	}
	if !foundSQLite {
		t.Error("Expected SQLite3 Database Config in doctor report")
	}
}

func TestScaffoldGoGinPostgresAndRedis(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "gin-pg-redis")

	fw, _ := config.FindFramework("gin")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:        "gin-pg-redis",
		TargetPath:  targetPath,
		Framework:   fw,
		Database:    db,
		Redis:       true,
		Docker:      true,
		NoInstall:   true,
		NoGit:       true,
		Interactive: false,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute failed: %v", err)
	}

	// 1. Verify go.mod contains both postgres gorm driver and go-redis
	modBytes, err := os.ReadFile(filepath.Join(targetPath, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	modContent := string(modBytes)
	if !strings.Contains(modContent, "gorm.io/driver/postgres") {
		t.Error("go.mod missing gorm.io/driver/postgres")
	}
	if !strings.Contains(modContent, "github.com/redis/go-redis/v9") {
		t.Error("go.mod missing github.com/redis/go-redis/v9")
	}

	// 2. Verify internal/database/redis.go
	redisGoPath := filepath.Join(targetPath, "internal", "database", "redis.go")
	if _, err := os.Stat(redisGoPath); os.IsNotExist(err) {
		t.Errorf("Expected helper '%s' was not generated", redisGoPath)
	}

	// 3. Verify docker-compose.yml
	composeBytes, err := os.ReadFile(filepath.Join(targetPath, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}
	composeContent := string(composeBytes)
	if !strings.Contains(composeContent, "image: postgres:16-alpine") {
		t.Error("docker-compose.yml missing postgres image")
	}
	if !strings.Contains(composeContent, "image: redis:7-alpine") {
		t.Error("docker-compose.yml missing redis image")
	}

	// 4. Scoped Doctor
	report, err := doctor.DiagnoseProject(targetPath)
	if err != nil {
		t.Fatalf("DiagnoseProject failed: %v", err)
	}

	foundPostgresDriver := false
	foundRedisDriver := false
	for _, check := range report.Checks {
		if check.Name == "PostgreSQL Driver" {
			foundPostgresDriver = true
		}
		if check.Name == "Redis Driver" {
			foundRedisDriver = true
		}
	}
	if !foundPostgresDriver {
		t.Error("Go Gin doctor report missing PostgreSQL Driver check")
	}
	if !foundRedisDriver {
		t.Error("Go Gin doctor report missing Redis Driver check")
	}
}

func TestScaffoldExpressMongoAndRedis(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "express-mongo-redis")

	fw, _ := config.FindFramework("express")
	db, _ := config.FindDatabase("mongo")

	cfg := &config.ProjectConfig{
		Name:        "express-mongo-redis",
		TargetPath:  targetPath,
		Framework:   fw,
		Database:    db,
		Redis:       true,
		Docker:      true,
		NoInstall:   true,
		NoGit:       true,
		Interactive: false,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute failed: %v", err)
	}

	// 1. Verify package.json contains mongodb / ioredis
	pkgBytes, err := os.ReadFile(filepath.Join(targetPath, "package.json"))
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	pkgContent := string(pkgBytes)
	if !strings.Contains(pkgContent, `"mongodb"`) && !strings.Contains(pkgContent, `"mongoose"`) {
		t.Error("package.json missing mongodb/mongoose dependency")
	}
	if !strings.Contains(pkgContent, `"ioredis"`) {
		t.Error("package.json missing ioredis dependency")
	}

	// 2. Verify src/lib/redis.ts
	redisTsPath := filepath.Join(targetPath, "src", "lib", "redis.ts")
	if _, err := os.Stat(redisTsPath); os.IsNotExist(err) {
		t.Errorf("Expected helper '%s' was not generated", redisTsPath)
	}

	// 3. Verify docker-compose.yml
	composeBytes, err := os.ReadFile(filepath.Join(targetPath, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}
	composeContent := string(composeBytes)
	if !strings.Contains(composeContent, "image: mongo:7-jammy") {
		t.Error("docker-compose.yml missing mongo:7-jammy image")
	}
	if !strings.Contains(composeContent, "image: redis:7-alpine") {
		t.Error("docker-compose.yml missing redis:7-alpine image")
	}

	// 4. Scoped Doctor
	report, err := doctor.DiagnoseProject(targetPath)
	if err != nil {
		t.Fatalf("DiagnoseProject failed: %v", err)
	}

	foundMongoDriver := false
	foundRedisDriver := false
	for _, check := range report.Checks {
		if check.Name == "MongoDB Driver" {
			foundMongoDriver = true
		}
		if check.Name == "Redis Driver" {
			foundRedisDriver = true
		}
	}
	if !foundMongoDriver {
		t.Error("Express doctor report missing MongoDB Driver check")
	}
	if !foundRedisDriver {
		t.Error("Express doctor report missing Redis Driver check")
	}
}
