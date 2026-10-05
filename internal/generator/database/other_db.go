package database

import (
	"fmt"

	"github.com/opeteer/ace/internal/config"
)

// InjectMySQL configures MySQL / MariaDB connection settings
func InjectMySQL(cfg *config.ProjectConfig) error {
	host := "127.0.0.1"
	if cfg.Docker {
		host = "db"
	}

	envConfig := fmt.Sprintf(`
# MySQL Database Configuration
DB_CONNECTION=mysql
DB_HOST=%s
DB_PORT=3306
DB_DATABASE=%s
DB_USERNAME=ace_user
DB_PASSWORD=secret
`, host, cfg.DBName())

	return appendToEnv(cfg.TargetPath, envConfig)
}

// InjectSQLite configures SQLite database settings
func InjectSQLite(cfg *config.ProjectConfig) error {
	envConfig := `
# SQLite Database Configuration
DB_CONNECTION=sqlite
DB_DATABASE=database.sqlite
`
	return appendToEnv(cfg.TargetPath, envConfig)
}

// InjectMongo configures MongoDB connection settings
func InjectMongo(cfg *config.ProjectConfig) error {
	host := "127.0.0.1"
	if cfg.Docker {
		host = "db"
	}

	envConfig := fmt.Sprintf(`
# MongoDB Configuration
MONGODB_URI=mongodb://ace_user:secret@%s:27017/%s?authSource=admin
`, host, cfg.DBName())

	return appendToEnv(cfg.TargetPath, envConfig)
}

// InjectRedis configures Redis connection settings
func InjectRedis(cfg *config.ProjectConfig) error {
	host := "127.0.0.1"
	if cfg.Docker {
		host = "redis"
	}

	envConfig := fmt.Sprintf(`
# Redis Cache Configuration
REDIS_HOST=%s
REDIS_PORT=6379
REDIS_PASSWORD=
`, host)

	return appendToEnv(cfg.TargetPath, envConfig)
}
