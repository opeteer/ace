package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/generator/cicd"
	"github.com/opeteer/ace/internal/generator/database"
	"github.com/opeteer/ace/internal/generator/devops"
	"github.com/opeteer/ace/internal/generator/framework"
	"github.com/opeteer/ace/internal/generator/proxy"
)

var (
	stepSuccess = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87")).Render("✔")
	stepName    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
)

// Engine manages the end-to-end scaffolding pipeline
type Engine struct {
	cfg *config.ProjectConfig
}

// NewEngine creates a new generator engine instance
func NewEngine(cfg *config.ProjectConfig) *Engine {
	return &Engine{cfg: cfg}
}

// Execute runs the scaffolding pipeline sequentially
func (e *Engine) Execute() error {
	// 1. Prepare target directory
	if err := e.prepareDirectory(); err != nil {
		return fmt.Errorf("failed to prepare directory: %w", err)
	}

	// 2. Scaffold Base Framework
	if err := e.scaffoldFramework(); err != nil {
		return fmt.Errorf("failed to scaffold framework: %w", err)
	}
	fmt.Printf("  %s %s\n", stepSuccess, stepName.Render(fmt.Sprintf("Scaffolded %s project (%s)", e.cfg.Framework.Name, e.cfg.Framework.Language)))

	// 3. Database Injection
	if e.cfg.Database.ID != "" {
		if err := database.InjectDatabase(e.cfg); err != nil {
			return fmt.Errorf("failed to configure database: %w", err)
		}
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render(fmt.Sprintf("Configured %s (%s)", e.cfg.Database.Name, e.cfg.Database.Paradigm)))
	}

	// 4. Docker & Compose
	if e.cfg.Docker {
		if err := devops.GenerateDocker(e.cfg); err != nil {
			return fmt.Errorf("failed to generate Dockerfile: %w", err)
		}
		if err := devops.GenerateCompose(e.cfg); err != nil {
			return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
		}
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Generated Dockerfile & docker-compose.yml"))
	}

	// 5. Reverse Proxy
	if e.cfg.Proxy == config.ProxyNginx {
		if err := proxy.GenerateNginx(e.cfg); err != nil {
			return fmt.Errorf("failed to generate Nginx config: %w", err)
		}
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Configured Nginx reverse proxy"))
	}

	// 6. CI/CD Workflows
	if e.cfg.CI == config.CIGitHub {
		if err := cicd.GenerateGitHubActions(e.cfg); err != nil {
			return fmt.Errorf("failed to generate GitHub Actions: %w", err)
		}
		fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Generated GitHub Actions CI workflow"))
	}

	// 7. Git Initialization
	if !e.cfg.NoGit {
		if err := e.initGit(); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Initialized Git repository"))
		}
	}

	// 8. Toolchain Integration Check (Host tools & Docker build readiness)
	e.checkBuildReadiness()

	return nil
}

func (e *Engine) checkBuildReadiness() {
	switch e.cfg.Framework.ID {
	case "laravel":
		if _, err := exec.LookPath("composer"); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Detected Composer on host (ready to run 'composer install')"))
		} else if e.cfg.Docker {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Containerized Composer & entrypoint pre-configured (ready for 'docker compose up --build')"))
		} else {
			fmt.Printf("  ! %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#E5A93C")).Render("Note: Composer is not on host. Install Composer or re-run with --docker for containerized builds."))
		}
	case "next":
		if _, err := exec.LookPath("npm"); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Detected Node.js & npm on host (ready for 'npm install')"))
		}
	case "fastapi":
		if _, err := exec.LookPath("python3"); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Detected Python on host (ready for virtualenv & pip)"))
		}
	case "fiber":
		if _, err := exec.LookPath("go"); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Detected Go toolchain on host (ready for 'go run')"))
		}
	}
}

func (e *Engine) prepareDirectory() error {
	target := e.cfg.TargetPath
	if stat, err := os.Stat(target); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("target path '%s' exists and is not a directory", target)
		}
		entries, err := os.ReadDir(target)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("target directory '%s' is not empty", target)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) scaffoldFramework() error {
	switch e.cfg.Framework.ID {
	case "laravel":
		return framework.ScaffoldLaravel(e.cfg)
	case "fastapi":
		return framework.ScaffoldFastAPI(e.cfg)
	case "next":
		return framework.ScaffoldNextJS(e.cfg)
	case "fiber":
		return framework.ScaffoldFiber(e.cfg)
	default:
		// Generic fallback generator for other frameworks
		return framework.ScaffoldGeneric(e.cfg)
	}
}

func (e *Engine) initGit() error {
	gitDir := filepath.Join(e.cfg.TargetPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = e.cfg.TargetPath
	_ = cmd.Run()

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = e.cfg.TargetPath
	_ = addCmd.Run()

	commitCmd := exec.Command("git", "commit", "-m", "chore: scaffold project with Ace CLI")
	commitCmd.Dir = e.cfg.TargetPath
	_ = commitCmd.Run()

	return nil
}
