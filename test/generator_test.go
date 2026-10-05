package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/generator"
)

func TestScaffoldLaravelPostgresFull(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-laravel-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "test-laravel")

	fw, _ := config.FindFramework("laravel")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:       "test-laravel",
		TargetPath: targetDir,
		Framework:  fw,
		Database:   db,
		Docker:     true,
		Proxy:      config.ProxyNginx,
		CI:         config.CIGitHub,
		NoGit:      false,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	// Verify framework files
	assertFileExists(t, filepath.Join(targetDir, "artisan"))
	assertFileExists(t, filepath.Join(targetDir, "composer.json"))
	assertFileExists(t, filepath.Join(targetDir, "bootstrap/app.php"))
	assertFileExists(t, filepath.Join(targetDir, "routes/web.php"))
	assertFileExists(t, filepath.Join(targetDir, "routes/api.php"))

	// Verify database configuration in .env
	assertFileContains(t, filepath.Join(targetDir, ".env"), "DB_CONNECTION=pgsql")
	assertFileContains(t, filepath.Join(targetDir, ".env"), "DB_HOST=db")
	assertFileContains(t, filepath.Join(targetDir, ".env"), "DB_DATABASE=test_laravel")

	// Verify Docker and Compose
	assertFileExists(t, filepath.Join(targetDir, "Dockerfile"))
	assertFileContains(t, filepath.Join(targetDir, "Dockerfile"), "php:8.3-fpm-alpine")
	assertFileExists(t, filepath.Join(targetDir, "docker-compose.yml"))
	assertFileContains(t, filepath.Join(targetDir, "docker-compose.yml"), "postgres:16-alpine")
	assertFileContains(t, filepath.Join(targetDir, "docker-compose.yml"), "nginx:alpine")

	// Verify Nginx
	assertFileExists(t, filepath.Join(targetDir, "docker/nginx/default.conf"))
	assertFileContains(t, filepath.Join(targetDir, "docker/nginx/default.conf"), "fastcgi_pass app:9000;")

	// Verify GitHub Actions CI
	assertFileExists(t, filepath.Join(targetDir, ".github/workflows/ci.yml"))
	assertFileContains(t, filepath.Join(targetDir, ".github/workflows/ci.yml"), "shivammathur/setup-php@v2")

	// Verify Git init
	assertDirExists(t, filepath.Join(targetDir, ".git"))
}

func TestScaffoldFastAPIPostgres(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-fastapi-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "test-fastapi")

	fw, _ := config.FindFramework("fastapi")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:       "test-fastapi",
		TargetPath: targetDir,
		Framework:  fw,
		Database:   db,
		Docker:     true,
		Proxy:      config.ProxyNginx,
		CI:         config.CIGitHub,
		NoGit:      true,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	// Verify FastAPI files
	assertFileExists(t, filepath.Join(targetDir, "app/main.py"))
	assertFileExists(t, filepath.Join(targetDir, "requirements.txt"))
	assertFileExists(t, filepath.Join(targetDir, "app/db/session.py"))
	assertFileExists(t, filepath.Join(targetDir, "app/models/item.py"))
	assertFileContains(t, filepath.Join(targetDir, "app/db/session.py"), "create_async_engine")

	// Verify requirements.txt has sqlalchemy & asyncpg
	assertFileContains(t, filepath.Join(targetDir, "requirements.txt"), "sqlalchemy")
	assertFileContains(t, filepath.Join(targetDir, "requirements.txt"), "asyncpg")

	// Verify Nginx reverse proxy
	assertFileExists(t, filepath.Join(targetDir, "docker/nginx/default.conf"))
	assertFileContains(t, filepath.Join(targetDir, "docker/nginx/default.conf"), "proxy_pass http://app_backend;")

	// Verify Git was skipped
	if _, err := os.Stat(filepath.Join(targetDir, ".git")); err == nil {
		t.Errorf("Expected .git to NOT exist with NoGit=true")
	}
}

func TestScaffoldNextJSPostgres(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-next-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "test-next")

	fw, _ := config.FindFramework("next")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:       "test-next",
		TargetPath: targetDir,
		Framework:  fw,
		Database:   db,
		Docker:     true,
		NoGit:      true,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	assertFileExists(t, filepath.Join(targetDir, "package.json"))
	assertFileExists(t, filepath.Join(targetDir, "app/layout.tsx"))
	assertFileExists(t, filepath.Join(targetDir, "app/page.tsx"))
	assertFileExists(t, filepath.Join(targetDir, "prisma/schema.prisma"))
	assertFileExists(t, filepath.Join(targetDir, "lib/prisma.ts"))
	assertFileContains(t, filepath.Join(targetDir, "prisma/schema.prisma"), "provider = \"postgresql\"")
	assertFileContains(t, filepath.Join(targetDir, "package.json"), "@prisma/client")
}

func TestScaffoldFiberPostgres(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-fiber-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "test-fiber")

	fw, _ := config.FindFramework("fiber")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:       "test-fiber",
		TargetPath: targetDir,
		Framework:  fw,
		Database:   db,
		Docker:     true,
		NoGit:      true,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	assertFileExists(t, filepath.Join(targetDir, "cmd/api/main.go"))
	assertFileExists(t, filepath.Join(targetDir, "internal/routes/routes.go"))
	assertFileExists(t, filepath.Join(targetDir, "internal/database/database.go"))
	assertFileContains(t, filepath.Join(targetDir, "internal/database/database.go"), "gorm.Open(postgres.Open(dsn)")
}

func TestBug03NestedPathDBName(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-nested-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "nested", "sub", "my-api")
	fw, _ := config.FindFramework("fastapi")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:       "nested/sub/my-api",
		TargetPath: targetDir,
		Framework:  fw,
		Database:   db,
		NoGit:      true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("cfg.Validate() failed: %v", err)
	}

	if cfg.Name != "my-api" {
		t.Errorf("Expected sanitized cfg.Name 'my-api', got '%s'", cfg.Name)
	}
	if cfg.DBName() != "my_api" {
		t.Errorf("Expected cfg.DBName() 'my_api', got '%s'", cfg.DBName())
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	// Verify no slashes in DATABASE_URL
	assertFileContains(t, filepath.Join(targetDir, ".env"), "DATABASE_URL=postgresql+asyncpg://ace_user:secret@127.0.0.1:5432/my_api")
	assertFileContains(t, filepath.Join(targetDir, "pyproject.toml"), `name = "my-api"`)
}

func TestBug04SpacesInProjectName(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-spaces-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "my test app")
	fw, _ := config.FindFramework("fiber")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:       "my test app",
		TargetPath: targetDir,
		Framework:  fw,
		Database:   db,
		NoGit:      true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("cfg.Validate() failed: %v", err)
	}

	if cfg.Name != "my-test-app" {
		t.Errorf("Expected slugified cfg.Name 'my-test-app', got '%s'", cfg.Name)
	}
	if cfg.DBName() != "my_test_app" {
		t.Errorf("Expected cfg.DBName() 'my_test_app', got '%s'", cfg.DBName())
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	// Verify go.mod has valid slugified module name without spaces
	assertFileContains(t, filepath.Join(targetDir, "go.mod"), "module my-test-app")
}

func TestBug06NginxAutoDocker(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-proxy-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "proxy-test")
	fw, _ := config.FindFramework("fastapi")

	cfg := &config.ProjectConfig{
		Name:       "proxy-test",
		TargetPath: targetDir,
		Framework:  fw,
		Proxy:      config.ProxyNginx,
		Docker:     false, // intentionally false to test auto-enable
		NoGit:      true,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("cfg.Validate() failed: %v", err)
	}

	if !cfg.Docker {
		t.Errorf("Expected cfg.Docker to be auto-enabled when Proxy is set")
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	assertFileExists(t, filepath.Join(targetDir, "Dockerfile"))
	assertFileExists(t, filepath.Join(targetDir, "docker-compose.yml"))
	assertFileExists(t, filepath.Join(targetDir, "docker/nginx/default.conf"))
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it does not", path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	stat, err := os.Stat(path)
	if os.IsNotExist(err) || !stat.IsDir() {
		t.Errorf("Expected directory %s to exist, but it does not", path)
	}
}

func assertFileContains(t *testing.T, path, sub string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Failed to read file %s: %v", path, err)
		return
	}
	if !strings.Contains(string(content), sub) {
		t.Errorf("Expected file %s to contain %q, but it did not", path, sub)
	}
}
