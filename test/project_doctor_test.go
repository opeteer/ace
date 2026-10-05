package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opeteer/ace/internal/doctor"
)

func TestDetectProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially empty: not a project
	if _, ok := doctor.DetectProject(tmpDir); ok {
		t.Error("Expected empty directory not to be detected as a project")
	}

	// 1. Laravel project
	laravelDir := filepath.Join(tmpDir, "my-laravel")
	_ = os.MkdirAll(laravelDir, 0755)
	_ = os.WriteFile(filepath.Join(laravelDir, "composer.json"), []byte(`{"name":"test/app"}`), 0644)
	_ = os.WriteFile(filepath.Join(laravelDir, "artisan"), []byte(`<?php echo "artisan";`), 0755)

	info, ok := doctor.DetectProject(laravelDir)
	if !ok {
		t.Fatal("Expected Laravel project to be detected")
	}
	if info.FrameworkID != "laravel" || info.Language != "PHP" {
		t.Errorf("Unexpected Laravel detection info: %+v", info)
	}

	// 2. FastAPI project
	fastapiDir := filepath.Join(tmpDir, "my-fastapi")
	_ = os.MkdirAll(fastapiDir, 0755)
	_ = os.WriteFile(filepath.Join(fastapiDir, "requirements.txt"), []byte("fastapi\nuvicorn\n"), 0644)

	info, ok = doctor.DetectProject(fastapiDir)
	if !ok {
		t.Fatal("Expected FastAPI project to be detected")
	}
	if info.FrameworkID != "fastapi" || info.Language != "Python" {
		t.Errorf("Unexpected FastAPI detection info: %+v", info)
	}

	// 3. Next.js project
	nextDir := filepath.Join(tmpDir, "my-next")
	_ = os.MkdirAll(nextDir, 0755)
	_ = os.WriteFile(filepath.Join(nextDir, "package.json"), []byte(`{"name":"next-app"}`), 0644)

	info, ok = doctor.DetectProject(nextDir)
	if !ok {
		t.Fatal("Expected Next.js project to be detected")
	}
	if info.FrameworkID != "next" {
		t.Errorf("Unexpected Next.js detection info: %+v", info)
	}

	// 4. Go project
	goDir := filepath.Join(tmpDir, "my-go")
	_ = os.MkdirAll(goDir, 0755)
	_ = os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module my-go\ngo 1.22\n"), 0644)

	info, ok = doctor.DetectProject(goDir)
	if !ok {
		t.Fatal("Expected Go project to be detected")
	}
	if info.FrameworkID != "fiber" || info.Language != "Go" {
		t.Errorf("Unexpected Go detection info: %+v", info)
	}
}

func TestDiagnoseProjectLaravelScoped(t *testing.T) {
	tmpDir := t.TempDir()
	laravelDir := filepath.Join(tmpDir, "test-laravel")
	_ = os.MkdirAll(filepath.Join(laravelDir, "vendor"), 0755)
	_ = os.WriteFile(filepath.Join(laravelDir, "composer.json"), []byte(`{"name":"test/laravel"}`), 0644)
	_ = os.WriteFile(filepath.Join(laravelDir, "artisan"), []byte("#!/usr/bin/env php"), 0755)
	_ = os.WriteFile(filepath.Join(laravelDir, "vendor", "autoload.php"), []byte("<?php"), 0644)
	_ = os.WriteFile(filepath.Join(laravelDir, ".env"), []byte("APP_NAME=test-laravel\nAPP_KEY=base64:dGVzdGtleQ==\nDB_CONNECTION=pgsql\nDB_HOST=127.0.0.1\nDB_PORT=5432\n"), 0644)

	report, err := doctor.DiagnoseProject(laravelDir)
	if err != nil {
		t.Fatalf("DiagnoseProject failed: %v", err)
	}

	if report.FrameworkID != "laravel" {
		t.Errorf("Expected framework_id 'laravel', got '%s'", report.FrameworkID)
	}

	// Verify that ONLY Laravel-relevant checks exist
	hasPHP := false
	hasComposer := false
	hasVendor := false
	hasEnv := false
	hasDB := false

	for _, check := range report.Checks {
		if strings.Contains(check.Name, "PHP") {
			hasPHP = true
		}
		if strings.Contains(check.Name, "Composer") {
			hasComposer = true
		}
		if strings.Contains(check.Name, "Vendor") {
			hasVendor = true
		}
		if strings.Contains(check.Name, "Environment") {
			hasEnv = true
		}
		if strings.Contains(check.Name, "Database") {
			hasDB = true
		}

		// Ensure unrelated languages are NEVER in the report
		if strings.Contains(check.Name, "Flutter") {
			t.Errorf("Unrelated check 'Flutter' found in Laravel project report!")
		}
		if strings.Contains(check.Name, "Rust") {
			t.Errorf("Unrelated check 'Rust' found in Laravel project report!")
		}
		if strings.Contains(check.Name, ".NET") || strings.Contains(check.Name, "dotnet") {
			t.Errorf("Unrelated check '.NET' found in Laravel project report!")
		}
	}

	if !hasPHP || !hasComposer || !hasVendor || !hasEnv || !hasDB {
		t.Errorf("Missing expected Laravel project checks: PHP=%v Composer=%v Vendor=%v Env=%v DB=%v",
			hasPHP, hasComposer, hasVendor, hasEnv, hasDB)
	}

	// Verify rendered report
	rendered := doctor.RenderProjectReport(report)
	if !strings.Contains(rendered, "Ace Project Doctor: Laravel 11 (PHP)") {
		t.Errorf("Rendered report missing title: %s", rendered)
	}
	if strings.Contains(rendered, "Flutter") {
		t.Errorf("Rendered report should not mention Flutter: %s", rendered)
	}
	if strings.Contains(rendered, "Rust") {
		t.Errorf("Rendered report should not mention Rust: %s", rendered)
	}
}

func TestDiagnoseFrameworkTarget(t *testing.T) {
	report, err := doctor.DiagnoseFramework("laravel")
	if err != nil {
		t.Fatalf("DiagnoseFramework('laravel') failed: %v", err)
	}

	for _, check := range report.Checks {
		if strings.Contains(check.Name, "Flutter") || strings.Contains(check.Name, "Rust") {
			t.Errorf("Unexpected tool in framework target check: %s", check.Name)
		}
	}

	reportFlutter, err := doctor.DiagnoseFramework("flutter")
	if err != nil {
		t.Fatalf("DiagnoseFramework('flutter') failed: %v", err)
	}
	foundFlutter := false
	for _, check := range reportFlutter.Checks {
		if strings.Contains(check.Name, "Flutter") {
			foundFlutter = true
		}
	}
	if !foundFlutter {
		t.Error("Expected Flutter check in flutter framework diagnosis")
	}
}
