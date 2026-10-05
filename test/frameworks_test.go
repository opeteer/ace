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

func TestAll19FrameworksScaffoldingAndScopedDoctor(t *testing.T) {
	tmpBase := t.TempDir()

	for _, fw := range config.SupportedFrameworks {
		fw := fw
		t.Run(fw.ID, func(t *testing.T) {
			projDir := filepath.Join(tmpBase, "proj-"+fw.ID)
			cfg := &config.ProjectConfig{
				Name:       "test-" + fw.ID,
				TargetPath: projDir,
				Framework:  fw,
				Database: config.DatabaseSpec{
					ID:          "postgres",
					Name:        "PostgreSQL 16+",
					Paradigm:    "Relational (SQL)",
					DefaultPort: 5432,
					DockerImage: "postgres:16-alpine",
					EnvPrefix:   "POSTGRES",
				},
				Docker:    true,
				NoInstall: true,
				NoGit:     true,
			}

			eng := generator.NewEngine(cfg)
			if err := eng.Execute(); err != nil {
				t.Fatalf("Scaffolding failed for %s: %v", fw.ID, err)
			}

			// 1. Verify Dockerfile and compose exist
			dockerfile := filepath.Join(projDir, "Dockerfile")
			if _, err := os.Stat(dockerfile); err != nil {
				t.Errorf("[%s] Expected Dockerfile to exist", fw.ID)
			}
			compose := filepath.Join(projDir, "docker-compose.yml")
			if _, err := os.Stat(compose); err != nil {
				t.Errorf("[%s] Expected docker-compose.yml to exist", fw.ID)
			}

			// 2. Verify framework-specific manifest file exists
			manifestFile := ""
			switch fw.ID {
			case "laravel":
				manifestFile = "composer.json"
			case "django":
				manifestFile = "manage.py"
			case "fastapi":
				manifestFile = "requirements.txt"
			case "fiber", "gin":
				manifestFile = "go.mod"
			case "axum":
				manifestFile = "Cargo.toml"
			case "flutter":
				manifestFile = "pubspec.yaml"
			case "aspnet":
				matches, _ := filepath.Glob(filepath.Join(projDir, "*.csproj"))
				if len(matches) == 0 {
					t.Errorf("[%s] Expected *.csproj file to exist", fw.ID)
				}
			case "springboot":
				manifestFile = "pom.xml"
			case "rails":
				manifestFile = "Gemfile"
			default:
				manifestFile = "package.json"
			}

			if manifestFile != "" {
				if _, err := os.Stat(filepath.Join(projDir, manifestFile)); err != nil {
					t.Errorf("[%s] Expected manifest '%s' to exist", fw.ID, manifestFile)
				}
			}

			// 3. Verify DetectProject accurately identifies the framework
			info, detected := doctor.DetectProject(projDir)
			if !detected {
				t.Fatalf("[%s] DetectProject failed to detect project", fw.ID)
			}
			if info.FrameworkID != fw.ID {
				t.Errorf("[%s] DetectProject returned framework '%s', expected '%s'", fw.ID, info.FrameworkID, fw.ID)
			}

			// 4. Verify DiagnoseProject produces a scoped report without unrelated toolchains
			report, err := doctor.DiagnoseProject(projDir)
			if err != nil {
				t.Fatalf("[%s] DiagnoseProject failed: %v", fw.ID, err)
			}

			if report.FrameworkID != fw.ID {
				t.Errorf("[%s] Report FrameworkID is '%s', expected '%s'", fw.ID, report.FrameworkID, fw.ID)
			}

			// Scoping audit: Ensure NO unrelated toolchains are in the check list
			for _, c := range report.Checks {
				checkName := strings.ToLower(c.Name)

				if fw.Language != "PHP" && strings.Contains(checkName, "php") {
					t.Errorf("[%s] Unrelated toolchain 'PHP' in report: %s", fw.ID, c.Name)
				}
				if fw.Language != "Python" && strings.Contains(checkName, "python") {
					t.Errorf("[%s] Unrelated toolchain 'Python' in report: %s", fw.ID, c.Name)
				}
				if fw.Language != "Rust" && (strings.Contains(checkName, "cargo") || strings.Contains(checkName, "rust")) {
					t.Errorf("[%s] Unrelated toolchain 'Rust' in report: %s", fw.ID, c.Name)
				}
				if fw.Language != "Dart" && strings.Contains(checkName, "flutter") {
					t.Errorf("[%s] Unrelated toolchain 'Flutter' in report: %s", fw.ID, c.Name)
				}
				if fw.Language != "C#" && strings.Contains(checkName, ".net") {
					t.Errorf("[%s] Unrelated toolchain '.NET' in report: %s", fw.ID, c.Name)
				}
				if !strings.Contains(fw.Language, "Java") && strings.Contains(checkName, "java") {
					t.Errorf("[%s] Unrelated toolchain 'Java' in report: %s", fw.ID, c.Name)
				}
			}

			// Verify RenderProjectReport outputs without error
			rendered := doctor.RenderProjectReport(report)
			if rendered == "" {
				t.Errorf("[%s] RenderProjectReport returned empty string", fw.ID)
			}
		})
	}
}
