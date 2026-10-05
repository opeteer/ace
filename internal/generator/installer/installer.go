package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/doctor"
)

var (
	stepSuccess = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87")).Render("✔")
	stepInfo    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D7D7")).Render("ℹ")
	stepWarn    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E5A93C")).Render("!")
	stepName    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	dimText     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

// Run coordinates automatic package installation and framework sync hooks
func Run(cfg *config.ProjectConfig) {
	if cfg.NoInstall {
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Skipped dependency installation (--no-install)"))
		return
	}

	switch cfg.Framework.ID {
	case "laravel":
		installLaravel(cfg)
	case "fastapi", "django":
		installPython(cfg)
	case "fiber", "gin":
		installGo(cfg)
	case "next", "nuxt", "sveltekit", "astro", "express", "nestjs", "vite-react", "vite-vue", "expo":
		installNode(cfg)
	case "axum":
		installRust(cfg)
	case "flutter":
		installFlutter(cfg)
	default:
		switch cfg.Framework.Language {
		case "Go":
			installGo(cfg)
		case "Python":
			installPython(cfg)
		case "TypeScript", "JavaScript":
			installNode(cfg)
		case "Rust":
			installRust(cfg)
		}
	}
}

// InstallDirectory inspects an existing project directory, auto-detects its ecosystem, and installs dependencies
func InstallDirectory(targetDir string) error {
	if targetDir == "" {
		targetDir = "."
	}
	targetDir = filepath.Clean(targetDir)

	stat, err := os.Stat(targetDir)
	if err != nil || !stat.IsDir() {
		return fmt.Errorf("directory '%s' does not exist or is not a directory", targetDir)
	}

	detected := false

	// Check Composer / PHP
	if fileExists(filepath.Join(targetDir, "composer.json")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir, Framework: config.FrameworkSpec{ID: "laravel"}}
		installLaravel(cfg)
	}

	// Check Node.js
	if fileExists(filepath.Join(targetDir, "package.json")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		if fileExists(filepath.Join(targetDir, "nuxt.config.ts")) || fileExists(filepath.Join(targetDir, "nuxt.config.js")) {
			cfg.Framework.ID = "nuxt"
		} else if fileExists(filepath.Join(targetDir, "svelte.config.js")) {
			cfg.Framework.ID = "sveltekit"
		} else if fileExists(filepath.Join(targetDir, "astro.config.mjs")) {
			cfg.Framework.ID = "astro"
		}
		installNode(cfg)
	}

	// Check Python
	if fileExists(filepath.Join(targetDir, "requirements.txt")) || fileExists(filepath.Join(targetDir, "pyproject.toml")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		installPython(cfg)
	}

	// Check Go
	if fileExists(filepath.Join(targetDir, "go.mod")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		installGo(cfg)
	}

	// Check Rust
	if fileExists(filepath.Join(targetDir, "Cargo.toml")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		installRust(cfg)
	}

	// Check Flutter
	if fileExists(filepath.Join(targetDir, "pubspec.yaml")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		installFlutter(cfg)
	}

	if !detected {
		return fmt.Errorf("no supported project manifest (package.json, composer.json, requirements.txt, go.mod, Cargo.toml, pubspec.yaml) found in '%s'", targetDir)
	}

	return nil
}

func installLaravel(cfg *config.ProjectConfig) {
	compBin, found := findComposerBinary()
	if !found {
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Composer not found on host. Downloading Composer..."))
		if err := doctor.InstallSpecificTool("composer"); err == nil {
			compBin, found = findComposerBinary()
		}
	}

	if found {
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Running 'composer install' for Laravel 11 (live logs enabled)..."))
		cmd := exec.Command(compBin, "install", "--no-interaction", "--prefer-dist", "--no-security-blocking")
		cmd.Dir = cfg.TargetPath
		if err := RunWithIdleTimeout(cmd, 35*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Installed Composer dependencies"))
			return
		}

		// Fallback without --no-security-blocking for older Composer versions
		cmd = exec.Command(compBin, "install", "--no-interaction", "--prefer-dist")
		cmd.Dir = cfg.TargetPath
		if err := RunWithIdleTimeout(cmd, 35*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Installed Composer dependencies"))
			return
		}

		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Host 'composer install' skipped or failed. Checking Docker fallback..."))
	}

	// Docker fallback: if Docker is available and enabled, run composer install inside container to populate vendor/
	if cfg.Docker && doctor.IsToolInstalled("docker") {
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Installing dependencies via Docker container..."))
		absTarget, err := filepath.Abs(cfg.TargetPath)
		if err == nil {
			cmd := exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/app", absTarget), "-w", "/app", "composer:latest", "composer", "install", "--no-interaction", "--prefer-dist")
			if err := RunWithIdleTimeout(cmd, 45*time.Second); err == nil {
				fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Installed Composer dependencies via containerized Composer"))
				return
			}
		}
	}

	if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Composer pre-configured (ready for 'docker compose up --build')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Composer not found on host. Run 'ace install composer' or use --docker."))
	}
}

func findComposerBinary() (string, bool) {
	if p, err := exec.LookPath("composer"); err == nil {
		return p, true
	}
	home, err := os.UserHomeDir()
	if err == nil {
		localComp := filepath.Join(home, ".local", "bin", "composer")
		if stat, err := os.Stat(localComp); err == nil && !stat.IsDir() {
			return localComp, true
		}
	}
	return "", false
}

func installNode(cfg *config.ProjectConfig) {
	pkgMgr := detectNodePkgMgr()
	if pkgMgr == "" {
		if cfg.Docker {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Node.js runtime pre-configured (ready for 'docker compose up')"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Node.js / npm not found on host. Run 'ace doctor' or use --docker."))
		}
		return
	}

	fmt.Printf("  %s %s\n", stepInfo, dimText.Render(fmt.Sprintf("Installing dependencies with %s...", pkgMgr)))
	cmd := exec.Command(pkgMgr, "install")
	cmd.Dir = cfg.TargetPath
	if err := cmd.Run(); err != nil {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render(fmt.Sprintf("%s install skipped or offline. Run '%s install' when ready.", pkgMgr, pkgMgr)))
		return
	}
	fmt.Printf("  %s %s\n", stepSuccess, stepName.Render(fmt.Sprintf("Installed Node.js dependencies (%s)", pkgMgr)))

	// Post-install sync hooks
	runNodeSyncHooks(cfg)
}

func runNodeSyncHooks(cfg *config.ProjectConfig) {
	switch cfg.Framework.ID {
	case "nuxt":
		cmd := exec.Command("npx", "nuxi", "prepare")
		cmd.Dir = cfg.TargetPath
		_ = cmd.Run()
	case "sveltekit":
		cmd := exec.Command("npx", "svelte-kit", "sync")
		cmd.Dir = cfg.TargetPath
		_ = cmd.Run()
	case "astro":
		cmd := exec.Command("npx", "astro", "sync")
		cmd.Dir = cfg.TargetPath
		_ = cmd.Run()
	}

	// Prisma schema sync
	prismaPath := filepath.Join(cfg.TargetPath, "prisma", "schema.prisma")
	if _, err := os.Stat(prismaPath); err == nil {
		cmd := exec.Command("npx", "prisma", "generate")
		cmd.Dir = cfg.TargetPath
		_ = cmd.Run()
	}
}

func installPython(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("python3"); err == nil {
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Configuring Python virtual environment (.venv)..."))
		venvCmd := exec.Command("python3", "-m", "venv", ".venv")
		venvCmd.Dir = cfg.TargetPath
		if err := venvCmd.Run(); err != nil {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Could not create .venv. Run 'python3 -m venv .venv'."))
			return
		}

		pipBin := filepath.Join(cfg.TargetPath, ".venv", "bin", "pip")
		if runtime.GOOS == "windows" {
			pipBin = filepath.Join(cfg.TargetPath, ".venv", "Scripts", "pip.exe")
		}

		reqFile := filepath.Join(cfg.TargetPath, "requirements.txt")
		if _, err := os.Stat(reqFile); err == nil && fileExists(pipBin) {
			pipCmd := exec.Command(pipBin, "install", "--disable-pip-version-check", "-r", "requirements.txt")
			pipCmd.Dir = cfg.TargetPath
			if err := pipCmd.Run(); err == nil {
				fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Created .venv and installed Python packages"))
				return
			}
		}
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Created Python virtual environment (.venv)"))
	} else if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Python pre-configured (ready for 'docker compose up')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Python 3 not found on host. Run 'ace doctor' or use --docker."))
	}
}

func installGo(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("go"); err == nil {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = cfg.TargetPath
		if err := cmd.Run(); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Synchronized Go modules (go mod tidy)"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("go mod tidy skipped or offline. Run 'go mod tidy' when online."))
		}
	} else if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Go build pre-configured (ready for 'docker compose up')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Go not found on host. Run 'ace doctor' or use --docker."))
	}
}

func installRust(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("cargo"); err == nil {
		cmd := exec.Command("cargo", "check")
		cmd.Dir = cfg.TargetPath
		if err := cmd.Run(); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Cargo verified dependencies (cargo check)"))
		}
	} else if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Rust build pre-configured (ready for 'docker compose up')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Cargo not found on host. Run 'ace doctor' or use --docker."))
	}
}

func installFlutter(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("flutter"); err == nil {
		cmd := exec.Command("flutter", "pub", "get")
		cmd.Dir = cfg.TargetPath
		if err := cmd.Run(); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Resolved Flutter packages (flutter pub get)"))
		}
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Flutter not found on host. Run 'ace doctor' to check SDK."))
	}
}

func detectNodePkgMgr() string {
	for _, mgr := range []string{"bun", "pnpm", "npm"} {
		if _, err := exec.LookPath(mgr); err == nil {
			return mgr
		}
	}
	return ""
}

func fileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}
