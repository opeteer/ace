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
	if cfg.Database.ID != "" {
		var err error
		switch cfg.Database.ID {
		case "postgres":
			err = InjectPostgres(cfg)
		case "mysql", "mariadb":
			err = InjectMySQL(cfg)
		case "mongo":
			err = InjectMongo(cfg)
		case "sqlite":
			err = InjectSQLite(cfg)
		case "redis":
			err = InjectRedis(cfg)
		default:
			err = injectGenericDB(cfg)
		}
		if err != nil {
			return err
		}

		host := "127.0.0.1"
		if cfg.Docker {
			host = "db"
		}
		if err := InjectDriverManifest(cfg, cfg.Database.ID, host); err != nil {
			return err
		}
	}

	if cfg.Redis && cfg.Database.ID != "redis" {
		if err := InjectCompanionRedis(cfg); err != nil {
			return err
		}
	}

	return nil
}

func appendToEnv(basePath, content string) error {
	envFile := filepath.Join(basePath, ".env")
	envExampleFile := filepath.Join(basePath, ".env.example")

	for _, file := range []string{envFile, envExampleFile} {
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		_, _ = f.WriteString("\n" + strings.TrimSpace(content) + "\n")
		_ = f.Close()
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
