package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/opeteer/ace/internal/config"
)

var (
	cardBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		MarginTop(1)

	successBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF87"))

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA"))

	valStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	cmdStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D7D7"))
)

// PrintSummary outputs a structured summary of the created project
func PrintSummary(cfg *config.ProjectConfig) {
	var body string
	body += fmt.Sprintf("%s Project %s created successfully!\n\n", successBadge.Render("✔"), keyStyle.Render(cfg.Name))
	body += fmt.Sprintf("  %s %s (%s)\n", keyStyle.Render("Framework:"), valStyle.Render(cfg.Framework.Name), valStyle.Render(cfg.Framework.Language))

	if cfg.Database.ID != "" {
		body += fmt.Sprintf("  %s %s [%s]\n", keyStyle.Render("Database: "), valStyle.Render(cfg.Database.Name), valStyle.Render(cfg.Database.Paradigm))
	} else {
		body += fmt.Sprintf("  %s %s\n", keyStyle.Render("Database: "), valStyle.Render("None"))
	}

	if cfg.Docker {
		body += fmt.Sprintf("  %s %s\n", keyStyle.Render("DevOps:   "), valStyle.Render("Dockerfile & docker-compose.yml"))
	}
	if cfg.Proxy != "" && cfg.Proxy != config.ProxyNone {
		body += fmt.Sprintf("  %s %s\n", keyStyle.Render("Proxy:    "), valStyle.Render(string(cfg.Proxy)))
	}
	if cfg.CI != "" && cfg.CI != config.CINone {
		body += fmt.Sprintf("  %s %s\n", keyStyle.Render("CI/CD:    "), valStyle.Render(string(cfg.CI)))
	}
	if !cfg.NoGit {
		body += fmt.Sprintf("  %s %s\n", keyStyle.Render("Git:      "), valStyle.Render("Initialized (.git)"))
	}

	body += fmt.Sprintf("\n%s\n", keyStyle.Render("Next Steps:"))
	body += fmt.Sprintf("  %s %s\n", cmdStyle.Render("cd"), cfg.TargetPath)

	if cfg.Docker {
		body += fmt.Sprintf("  %s\n", cmdStyle.Render("docker compose up -d"))
		if cfg.Framework.ID == "laravel" {
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("docker compose exec app php artisan migrate"))
		}
	} else {
		// Native startup commands
		switch cfg.Framework.ID {
		case "laravel":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("composer install && php artisan serve"))
		case "fastapi":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("python3 -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt\n  uvicorn app.main:app --reload"))
		case "django":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("python3 -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt\n  python manage.py runserver"))
		case "next", "nuxt", "sveltekit", "astro", "vite-react", "vite-vue":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("npm install && npm run dev"))
		case "express":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("npm install && npm run dev"))
		case "nestjs":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("npm install && npm run start:dev"))
		case "fiber":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("go run cmd/api/main.go"))
		case "gin":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("go run main.go"))
		case "axum":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("cargo run"))
		case "rails":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("bundle install && rails s"))
		case "springboot":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("./mvnw spring-boot:run"))
		case "aspnet":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("dotnet run"))
		case "flutter":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("flutter pub get && flutter run"))
		case "expo":
			body += fmt.Sprintf("  %s\n", cmdStyle.Render("npx expo start"))
		default:
			switch cfg.Framework.Language {
			case "TypeScript", "JavaScript":
				body += fmt.Sprintf("  %s\n", cmdStyle.Render("npm install && npm start"))
			case "Python":
				body += fmt.Sprintf("  %s\n", cmdStyle.Render("python3 -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt"))
			case "Dart":
				body += fmt.Sprintf("  %s\n", cmdStyle.Render("flutter pub get && flutter run"))
			case "Go":
				body += fmt.Sprintf("  %s\n", cmdStyle.Render("go run main.go"))
			case "Rust":
				body += fmt.Sprintf("  %s\n", cmdStyle.Render("cargo run"))
			case "Ruby":
				body += fmt.Sprintf("  %s\n", cmdStyle.Render("bundle install && rails s"))
			default:
				body += fmt.Sprintf("  %s\n", cmdStyle.Render("Check README.md for instructions"))
			}
		}
	}

	fmt.Println(cardBox.Render(body))
}
