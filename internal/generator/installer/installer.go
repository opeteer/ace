package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	case "aspnet":
		installDotnet(cfg)
	case "springboot":
		installJava(cfg)
	case "rails":
		installRails(cfg)
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
		case "C#":
			installDotnet(cfg)
		case "Java", "Java/Kotlin":
			installJava(cfg)
		case "Ruby":
			installRails(cfg)
		case "Dart":
			installFlutter(cfg)
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

	// Check .NET
	if matches, _ := filepath.Glob(filepath.Join(targetDir, "*.csproj")); len(matches) > 0 {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		installDotnet(cfg)
	}

	// Check Java / Maven
	if fileExists(filepath.Join(targetDir, "pom.xml")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		installJava(cfg)
	}

	// Check Ruby / Bundler
	if fileExists(filepath.Join(targetDir, "Gemfile")) {
		detected = true
		cfg := &config.ProjectConfig{TargetPath: targetDir}
		installRails(cfg)
	}

	if !detected {
		return fmt.Errorf("no supported project manifest found in '%s'", targetDir)
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
		cmd := exec.Command(compBin, "install", "--no-interaction", "--prefer-dist", "--no-security-blocking", "-v")
		cmd.Dir = cfg.TargetPath
		if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Installed Composer dependencies"))
			return
		}

		// Fallback without --no-security-blocking for older Composer versions
		cmd = exec.Command(compBin, "install", "--no-interaction", "--prefer-dist", "-v")
		cmd.Dir = cfg.TargetPath
		if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Installed Composer dependencies"))
			return
		}

		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Host 'composer install' skipped or failed. Checking Docker fallback..."))
	}

	// Docker fallback: if Docker is available, enabled, and daemon is actively running
	if cfg.Docker && isDockerDaemonRunning() {
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Installing dependencies via Docker container..."))
		absTarget, err := filepath.Abs(cfg.TargetPath)
		if err == nil {
			cmd := exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/app", absTarget), "-w", "/app", "composer:latest", "composer", "install", "--no-interaction", "--prefer-dist")
			if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
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

func isDockerDaemonRunning() bool {
	if !doctor.IsToolInstalled("docker") {
		return false
	}
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
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

	absTarget, err := filepath.Abs(cfg.TargetPath)
	if err != nil {
		absTarget = cfg.TargetPath
	}

	fwName := cfg.Framework.Name
	if fwName == "" {
		fwName = "Node.js"
	}

	fmt.Printf("  %s %s\n", stepInfo, dimText.Render(fmt.Sprintf("Running '%s install' for %s (live logs enabled)...", pkgMgr, fwName)))
	cmd := exec.Command(pkgMgr, "install")
	cmd.Dir = absTarget
	if err := RunWithIdleTimeout(cmd, 120*time.Second); err != nil {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render(fmt.Sprintf("%s install skipped or offline. Run '%s install' when ready.", pkgMgr, pkgMgr)))
		return
	}
	fmt.Printf("  %s %s\n", stepSuccess, stepName.Render(fmt.Sprintf("Installed Node.js dependencies (%s)", pkgMgr)))

	// Post-install sync hooks
	runNodeSyncHooks(cfg)
}

func runNodeSyncHooks(cfg *config.ProjectConfig) {
	absTarget, _ := filepath.Abs(cfg.TargetPath)

	switch cfg.Framework.ID {
	case "nuxt":
		cmd := exec.Command("npx", "nuxi", "prepare")
		cmd.Dir = absTarget
		_ = cmd.Run()
	case "sveltekit":
		cmd := exec.Command("npx", "svelte-kit", "sync")
		cmd.Dir = absTarget
		_ = cmd.Run()
	case "astro":
		cmd := exec.Command("npx", "astro", "sync")
		cmd.Dir = absTarget
		_ = cmd.Run()
	}

	// Prisma schema sync
	prismaPath := filepath.Join(absTarget, "prisma", "schema.prisma")
	if _, err := os.Stat(prismaPath); err == nil {
		cmd := exec.Command("npx", "prisma", "generate")
		cmd.Dir = absTarget
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

		absTarget, err := filepath.Abs(cfg.TargetPath)
		if err != nil {
			absTarget = cfg.TargetPath
		}

		pipBin := filepath.Join(absTarget, ".venv", "bin", "pip")
		pyBin := filepath.Join(absTarget, ".venv", "bin", "python")
		if runtime.GOOS == "windows" {
			pipBin = filepath.Join(absTarget, ".venv", "Scripts", "pip.exe")
			pyBin = filepath.Join(absTarget, ".venv", "Scripts", "python.exe")
		}

		reqFile := filepath.Join(cfg.TargetPath, "requirements.txt")
		if _, err := os.Stat(reqFile); err == nil && fileExists(pipBin) {
			fwName := cfg.Framework.Name
			if fwName == "" {
				fwName = "Python"
			}
			fmt.Printf("  %s %s\n", stepInfo, dimText.Render(fmt.Sprintf("Running 'pip install' for %s (live logs enabled)...", fwName)))
			pipCmd := exec.Command(pipBin, "install", "--disable-pip-version-check", "-r", "requirements.txt")
			pipCmd.Dir = cfg.TargetPath
			if err := RunWithIdleTimeout(pipCmd, 120*time.Second); err == nil {
				// Verify installed version
				verStr := ""
				if cfg.Framework.ID == "django" {
					verCmd := exec.Command(pyBin, "-m", "django", "--version")
					if out, vErr := verCmd.Output(); vErr == nil {
						verStr = fmt.Sprintf(" (Django %s)", strings.TrimSpace(string(out)))
					}
				} else if cfg.Framework.ID == "fastapi" {
					verStr = " (FastAPI)"
				}
				fmt.Printf("  %s %s\n", stepSuccess, stepName.Render(fmt.Sprintf("Created .venv and installed Python packages%s", verStr)))
				return
			}
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("pip install skipped or failed. Run 'pip install -r requirements.txt' when ready."))
			return
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
		absTarget, _ := filepath.Abs(cfg.TargetPath)
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Running 'go mod tidy' (live logs enabled)..."))
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = absTarget
		if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
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
		absTarget, _ := filepath.Abs(cfg.TargetPath)
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Running 'cargo check' (live logs enabled)..."))
		cmd := exec.Command("cargo", "check")
		cmd.Dir = absTarget
		if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Cargo verified dependencies (cargo check)"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("cargo check skipped or offline. Run 'cargo check' when online."))
		}
	} else if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Rust build pre-configured (ready for 'docker compose up')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Cargo not found on host. Run 'ace doctor' or use --docker."))
	}
}

func installFlutter(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("flutter"); err == nil {
		absTarget, _ := filepath.Abs(cfg.TargetPath)
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Running 'flutter pub get' (live logs enabled)..."))
		cmd := exec.Command("flutter", "pub", "get")
		cmd.Dir = absTarget
		if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Resolved Flutter packages (flutter pub get)"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("flutter pub get skipped or offline."))
		}
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Flutter not found on host. Run 'ace doctor' to check SDK."))
	}
}

func installDotnet(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("dotnet"); err == nil {
		absTarget, _ := filepath.Abs(cfg.TargetPath)
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Running 'dotnet restore' (live logs enabled)..."))
		cmd := exec.Command("dotnet", "restore")
		cmd.Dir = absTarget
		if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Restored .NET packages (dotnet restore)"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("dotnet restore skipped or offline. Run 'dotnet restore' when ready."))
		}
	} else if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized .NET build pre-configured (ready for 'docker compose up')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render(".NET SDK not found on host. Run 'ace doctor' or use --docker."))
	}
}

func installJava(cfg *config.ProjectConfig) {
	mvnBin := "mvn"
	absTarget, _ := filepath.Abs(cfg.TargetPath)
	if fileExists(filepath.Join(absTarget, "mvnw")) {
		mvnBin = "./mvnw"
	} else if _, err := exec.LookPath("mvn"); err != nil {
		if cfg.Docker {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Maven build pre-configured (ready for 'docker compose up')"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Maven / Java not found on host. Run 'ace doctor' or use --docker."))
		}
		return
	}

	fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Resolving Maven dependencies (live logs enabled)..."))
	cmd := exec.Command(mvnBin, "dependency:resolve", "-q")
	cmd.Dir = absTarget
	if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Resolved Maven dependencies"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Maven dependency resolution skipped. Run 'mvn compile' when ready."))
	}
}

func installRails(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("bundle"); err == nil {
		absTarget, _ := filepath.Abs(cfg.TargetPath)
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Running 'bundle install' for Ruby on Rails (live logs enabled)..."))
		cmd := exec.Command("bundle", "install")
		cmd.Dir = absTarget
		if err := RunWithIdleTimeout(cmd, 120*time.Second); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Installed Bundler gems"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("bundle install skipped or offline. Run 'bundle install' when ready."))
		}
	} else if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Ruby pre-configured (ready for 'docker compose up')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Bundler not found on host. Run 'gem install bundler' or use --docker."))
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
