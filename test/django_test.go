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

func TestScaffoldDjangoPostgresFull(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "my-django-app")

	fw, _ := config.FindFramework("django")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:        "my-django-app",
		TargetPath:  targetPath,
		Framework:   fw,
		Database:    db,
		Docker:      true,
		Proxy:       config.ProxyNginx,
		CI:          config.CIGitHub,
		NoInstall:   true,
		NoGit:       true,
		Interactive: false,
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute failed for Django: %v", err)
	}

	// 1. Verify Django project layout
	expectedFiles := []string{
		"manage.py",
		"requirements.txt",
		"pyproject.toml",
		"my_django_app/__init__.py",
		"my_django_app/settings.py",
		"my_django_app/urls.py",
		"my_django_app/views.py",
		"my_django_app/wsgi.py",
		"my_django_app/asgi.py",
		"Dockerfile",
		"docker-compose.yml",
		"docker/nginx/default.conf",
		".env",
		".env.example",
	}

	for _, f := range expectedFiles {
		p := filepath.Join(targetPath, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("Expected file '%s' was not generated", f)
		}
	}

	// 2. Verify requirements.txt
	reqBytes, err := os.ReadFile(filepath.Join(targetPath, "requirements.txt"))
	if err != nil {
		t.Fatalf("Failed to read requirements.txt: %v", err)
	}
	reqContent := string(reqBytes)
	if !strings.Contains(reqContent, "Django") {
		t.Error("requirements.txt missing Django")
	}
	if !strings.Contains(reqContent, "psycopg") {
		t.Error("requirements.txt missing psycopg")
	}
	if !strings.Contains(reqContent, "gunicorn") {
		t.Error("requirements.txt missing gunicorn")
	}

	// 3. Verify .env database configuration
	envBytes, err := os.ReadFile(filepath.Join(targetPath, ".env"))
	if err != nil {
		t.Fatalf("Failed to read .env: %v", err)
	}
	envContent := string(envBytes)
	if !strings.Contains(envContent, "DB_ENGINE=django.db.backends.postgresql") {
		t.Error(".env missing DB_ENGINE=django.db.backends.postgresql")
	}
	if !strings.Contains(envContent, "DB_NAME=my_django_app") {
		t.Error(".env missing DB_NAME=my_django_app")
	}

	// 4. Verify Dockerfile
	dockerBytes, err := os.ReadFile(filepath.Join(targetPath, "Dockerfile"))
	if err != nil {
		t.Fatalf("Failed to read Dockerfile: %v", err)
	}
	dockerContent := string(dockerBytes)
	if !strings.Contains(dockerContent, "gunicorn") {
		t.Error("Dockerfile missing gunicorn CMD")
	}
	if !strings.Contains(dockerContent, "my_django_app.wsgi:application") {
		t.Error("Dockerfile missing WSGI application target")
	}

	// 5. Verify Project Doctor Scoping
	report, err := doctor.DiagnoseProject(targetPath)
	if err != nil {
		t.Fatalf("DiagnoseProject failed for Django: %v", err)
	}

	if report.Framework != "Django 5" {
		t.Errorf("Expected framework 'Django 5', got '%s'", report.Framework)
	}

	// Check that unrelated tools do not appear
	for _, c := range report.Checks {
		if strings.Contains(c.Name, "Flutter") || strings.Contains(c.Name, "Rust") || strings.Contains(c.Name, "Composer") {
			t.Errorf("Unrelated check '%s' found in Django project report!", c.Name)
		}
	}

	// Verify rendered report
	rendered := doctor.RenderProjectReport(report)
	if !strings.Contains(rendered, "Django 5") {
		t.Errorf("Rendered report missing Django 5 title: %s", rendered)
	}
	if strings.Contains(rendered, "Flutter") {
		t.Error("Rendered report contains Flutter")
	}
}
