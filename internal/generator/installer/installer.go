package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/opeteer/ace/internal/config"
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

func installLaravel(cfg *config.ProjectConfig) {
	if _, err := exec.LookPath("composer"); err == nil {
		fmt.Printf("  %s %s\n", stepInfo, dimText.Render("Running 'composer install' for Laravel 11..."))
		cmd := exec.Command("composer", "install", "--no-interaction", "--prefer-dist", "--no-progress")
		cmd.Dir = cfg.TargetPath
		if err := cmd.Run(); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Installed Composer dependencies"))
		} else {
			fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Composer install skipped or offline. Run 'composer install'."))
		}
	} else if cfg.Docker {
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Composer pre-configured (ready for 'docker compose up --build')"))
	} else {
		fmt.Printf("  %s %s\n", stepWarn, dimText.Render("Composer not found on host. Run 'ace doctor' or use --docker."))
	}
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
