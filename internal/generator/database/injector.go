package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// InjectDatabase routes database injection based on engine and framework
func InjectDatabase(cfg *config.ProjectConfig) error {
	switch cfg.Database.ID {
	case "postgres":
		return InjectPostgres(cfg)
	case "mysql", "mariadb":
		return InjectMySQL(cfg)
	case "mongo":
		return InjectMongo(cfg)
	case "sqlite":
		return InjectSQLite(cfg)
	case "redis":
		return InjectRedis(cfg)
	default:
		return injectGenericDB(cfg)
	}
}

func appendToEnv(basePath, content string) error {
	envFile := filepath.Join(basePath, ".env")
	envExampleFile := filepath.Join(basePath, ".env.example")

	for _, file := range []string{envFile, envExampleFile} {
		if _, err := os.Stat(file); err == nil {
			f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			_, _ = f.WriteString("\n" + strings.TrimSpace(content) + "\n")
			_ = f.Close()
		}
	}
	return nil
}

func injectGenericDB(cfg *config.ProjectConfig) error {
	host := "127.0.0.1"
	if cfg.Docker {
		host = "db"
	}
	envStr := fmt.Sprintf(`# Database Configuration (%s)
DB_TYPE=%s
DB_HOST=%s
DB_PORT=%d
DB_NAME=%s
DB_USER=ace_user
DB_PASSWORD=secret
`, cfg.Database.Name, cfg.Database.ID, host, cfg.Database.DefaultPort, cfg.DBName())

	return appendToEnv(cfg.TargetPath, envStr)
}
