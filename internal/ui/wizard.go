package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/opeteer/ace/internal/config"
)

// RunWizard prompts the user interactively to collect ProjectConfig
func RunWizard(initialName string) (*config.ProjectConfig, error) {
	var (
		projectName    = initialName
		selectedCat    string
		selectedFwID   string
		selectedDbID   string
		enableDocker   bool
		selectedProxy  string
		selectedCI     string
		initGit        = true
	)

	// Step 1: Project Name
	if strings.TrimSpace(projectName) == "" {
		err := huh.NewInput().
			Title("What is your project name?").
			Placeholder("my-app").
			Value(&projectName).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("project name cannot be empty")
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
	}

	// Categories
	catOptions := []huh.Option[string]{
		huh.NewOption("Fullstack / SSR (Next.js, Laravel, Nuxt, Django, Rails...)", string(config.CategoryFullstack)),
		huh.NewOption("Backend API (FastAPI, Go Fiber/Gin, Express, NestJS, Spring Boot...)", string(config.CategoryBackend)),
		huh.NewOption("Frontend SPA (React Vite, Vue Vite...)", string(config.CategoryFrontend)),
		huh.NewOption("Mobile Cross-Platform (Flutter, React Native/Expo)", string(config.CategoryMobile)),
	}

	err := huh.NewSelect[string]().
		Title("Select Project Category").
		Options(catOptions...).
		Value(&selectedCat).
		Run()
	if err != nil {
		return nil, err
	}

	// Filter frameworks by selected category
	var fwOptions []huh.Option[string]
	for _, f := range config.SupportedFrameworks {
		if string(f.Category) == selectedCat {
			label := fmt.Sprintf("%-20s (%s) - %s", f.Name, f.Language, f.Description)
			fwOptions = append(fwOptions, huh.NewOption(label, f.ID))
		}
	}

	err = huh.NewSelect[string]().
		Title("Select Framework").
		Options(fwOptions...).
		Value(&selectedFwID).
		Run()
	if err != nil {
		return nil, err
	}

	// Step 3: Database
	var dbOptions []huh.Option[string]
	dbOptions = append(dbOptions, huh.NewOption("None (Stateless / In-Memory)", "none"))
	for _, d := range config.SupportedDatabases {
		label := fmt.Sprintf("%-20s [%s]", d.Name, d.Paradigm)
		dbOptions = append(dbOptions, huh.NewOption(label, d.ID))
	}

	err = huh.NewSelect[string]().
		Title("Select Database").
		Options(dbOptions...).
		Value(&selectedDbID).
		Run()
	if err != nil {
		return nil, err
	}

	// Step 4: Add-ons & DevOps Form
	proxyOptions := []huh.Option[string]{
		huh.NewOption("None (Direct app port)", string(config.ProxyNone)),
		huh.NewOption("Nginx (Reverse proxy + security headers)", string(config.ProxyNginx)),
		huh.NewOption("Caddy (Automatic HTTPS)", string(config.ProxyCaddy)),
	}

	ciOptions := []huh.Option[string]{
		huh.NewOption("None", string(config.CINone)),
		huh.NewOption("GitHub Actions (CI matrix & container build)", string(config.CIGitHub)),
		huh.NewOption("GitLab CI (.gitlab-ci.yml)", string(config.CIGitLab)),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Include Docker & Docker Compose?").
				Description("Generates multi-stage Dockerfile and orchestrated docker-compose.yml").
				Value(&enableDocker),

			huh.NewSelect[string]().
				Title("Reverse Proxy / Gateway").
				Options(proxyOptions...).
				Value(&selectedProxy),

			huh.NewSelect[string]().
				Title("CI/CD Pipeline").
				Options(ciOptions...).
				Value(&selectedCI),

			huh.NewConfirm().
				Title("Initialize Git repository?").
				Value(&initGit),
		),
	)

	err = form.Run()
	if err != nil {
		return nil, err
	}

	fw, _ := config.FindFramework(selectedFwID)
	var db config.DatabaseSpec
	if selectedDbID != "none" {
		db, _ = config.FindDatabase(selectedDbID)
	}

	cfg := &config.ProjectConfig{
		Name:        strings.TrimSpace(projectName),
		TargetPath:  strings.TrimSpace(projectName),
		Framework:   fw,
		Database:    db,
		Docker:      enableDocker,
		Proxy:       config.ProxyType(selectedProxy),
		Protocol:    config.ProtocolREST,
		CI:          config.CIType(selectedCI),
		NoGit:       !initGit,
		Interactive: true,
	}

	return cfg, nil
}
