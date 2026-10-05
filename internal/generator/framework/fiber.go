package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldFiber generates an idiomatic Go Fiber REST API microservice structure
func ScaffoldFiber(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"cmd/api",
		"internal/config",
		"internal/handlers",
		"internal/routes",
		"internal/database",
		"pkg/logger",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"go.mod": fmt.Sprintf(`module %s

go 1.22

require (
	github.com/gofiber/fiber/v2 v2.52.5
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
)
`, cfg.Name),

		"cmd/api/main.go": fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"%s/internal/routes"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := fiber.New(fiber.Config{
		AppName: "%s",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	routes.RegisterRoutes(app)

	log.Printf("Starting server on port %%s...", port)
	if err := app.Listen(fmt.Sprintf(":%%s", port)); err != nil {
		log.Fatalf("Server failed: %%v", err)
	}
}
`, cfg.Name, cfg.Name),

		"internal/handlers/health.go": "package handlers\n\nimport (\n\t\"time\"\n\n\t\"github.com/gofiber/fiber/v2\"\n)\n\ntype HealthResponse struct {\n\tStatus    string `json:\"status\"`\n\tTimestamp string `json:\"timestamp\"`\n}\n\nfunc HealthCheck(c *fiber.Ctx) error {\n\treturn c.JSON(HealthResponse{\n\t\tStatus:    \"healthy\",\n\t\tTimestamp: time.Now().UTC().Format(time.RFC3339),\n\t})\n}\n",

		"internal/routes/routes.go": fmt.Sprintf(`package routes

import (
	"github.com/gofiber/fiber/v2"
	"%s/internal/handlers"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	api.Get("/health", handlers.HealthCheck)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":           "%s",
			"status":        "online",
			"scaffolded_by": "Ace CLI",
		})
	})
}
`, cfg.Name, cfg.Name),

		".gitignore": `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
vendor/

# Environment files
.env
.env.local
`,

		".env.example": fmt.Sprintf(`APP_NAME=%s
PORT=8080
APP_ENV=development
`, cfg.Name),
	}

	for relPath, content := range files {
		targetFile := filepath.Join(base, relPath)
		if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
			return err
		}
	}

	envExample, err := os.ReadFile(filepath.Join(base, ".env.example"))
	if err == nil {
		_ = os.WriteFile(filepath.Join(base, ".env"), envExample, 0644)
	}

	return nil
}
