package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldGin generates a production-grade Go Gin microservice
func ScaffoldGin(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"cmd/api",
		"internal/handlers",
		"internal/routes",
		"internal/config",
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
	github.com/gin-gonic/gin v1.10.0
	github.com/joho/godotenv v1.5.1
)
`, cfg.Name),

		"cmd/api/main.go": fmt.Sprintf(`package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"%s/internal/routes"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "%d"
	}

	r := gin.Default()
	routes.RegisterRoutes(r)

	log.Printf(">> Gin server running on http://localhost:%%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %%v", err)
	}
}
`, cfg.Name, cfg.Framework.DefaultPort),

		"internal/routes/routes.go": fmt.Sprintf(`package routes

import (
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"app":           "%s",
			"framework":     "Go Gin",
			"status":        "online",
			"scaffolded_by": "Ace CLI",
		})
	})

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})
	}
}
`, cfg.Name),

		".env.example": fmt.Sprintf(`PORT=%d
ENV=development
`, cfg.Framework.DefaultPort),

		".gitignore": `bin/
*.out
*.exe
.env
.env.backup
.DS_Store
`,
	}

	for relPath, content := range files {
		targetFile := filepath.Join(base, relPath)
		_ = os.MkdirAll(filepath.Dir(targetFile), 0755)
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
