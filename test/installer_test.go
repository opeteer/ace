package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/generator"
	"github.com/opeteer/ace/internal/generator/installer"
)

func TestInstallerSkipNoInstall(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-no-install-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetDir := filepath.Join(tempDir, "test-fastapi-no-install")
	fw, _ := config.FindFramework("fastapi")
	db, _ := config.FindDatabase("postgres")

	cfg := &config.ProjectConfig{
		Name:       "test-fastapi-no-install",
		TargetPath: targetDir,
		Framework:  fw,
		Database:   db,
		NoGit:      true,
		NoInstall:  true, // explicitly skip dependency installation
	}

	engine := generator.NewEngine(cfg)
	if err := engine.Execute(); err != nil {
		t.Fatalf("Engine.Execute() failed: %v", err)
	}

	// Verify project was scaffolded
	assertFileExists(t, filepath.Join(targetDir, "requirements.txt"))

	// Verify .venv was NOT created because NoInstall was true
	venvPath := filepath.Join(targetDir, ".venv")
	if _, err := os.Stat(venvPath); err == nil {
		t.Errorf("Expected .venv to NOT exist when NoInstall=true")
	}
}

func TestInstallDirectoryNoManifest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ace-test-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = installer.InstallDirectory(tempDir)
	if err == nil {
		t.Fatal("Expected error when running InstallDirectory on directory with no manifest")
	}
}
