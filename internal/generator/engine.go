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
	"github.com/opeteer/ace/internal/generator/installer"
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
	if e.cfg.Database.ID != "" || e.cfg.Redis {
		if err := database.InjectDatabase(e.cfg); err != nil {
			return fmt.Errorf("failed to configure database: %w", err)
		}
		if e.cfg.Database.ID != "" {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render(fmt.Sprintf("Configured %s (%s)", e.cfg.Database.Name, e.cfg.Database.Paradigm)))
		}
		if e.cfg.Redis && e.cfg.Database.ID != "redis" {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Configured companion Redis (Key-Value / Cache)"))
		}
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

	// 7. Automated Dependency Provisioning & Sync Hooks
	installer.Run(e.cfg)

	// 8. Git Initialization
	if !e.cfg.NoGit {
		if err := e.initGit(); err == nil {
			fmt.Printf("  %s %s\n", stepSuccess, stepName.Render("Initialized Git repository"))
		}
	}

	return nil
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
	case "django":
		return framework.ScaffoldDjango(e.cfg)
	case "next":
		return framework.ScaffoldNextJS(e.cfg)
	case "nuxt":
		return framework.ScaffoldNuxt(e.cfg)
	case "sveltekit":
		return framework.ScaffoldSvelteKit(e.cfg)
	case "astro":
		return framework.ScaffoldAstro(e.cfg)
	case "express":
		return framework.ScaffoldExpress(e.cfg)
	case "nestjs":
		return framework.ScaffoldNestJS(e.cfg)
	case "vite-react":
		return framework.ScaffoldReact(e.cfg)
	case "vite-vue":
		return framework.ScaffoldVue(e.cfg)
	case "fiber":
		return framework.ScaffoldFiber(e.cfg)
	case "gin":
		return framework.ScaffoldGin(e.cfg)
	case "axum":
		return framework.ScaffoldAxum(e.cfg)
	case "aspnet":
		return framework.ScaffoldAspNet(e.cfg)
	case "springboot":
		return framework.ScaffoldSpringBoot(e.cfg)
	case "flutter":
		return framework.ScaffoldFlutter(e.cfg)
	case "expo":
		return framework.ScaffoldExpo(e.cfg)
	case "rails":
		return framework.ScaffoldRails(e.cfg)
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
