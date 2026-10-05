package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// InjectMySQL configures MySQL / MariaDB connection settings and drivers
func InjectMySQL(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath
	host := "127.0.0.1"
	if cfg.Docker {
		host = "db"
	}

	envConfig := fmt.Sprintf(`
# MySQL / MariaDB Database Configuration
DB_CONNECTION=mysql
DB_ENGINE=django.db.backends.mysql
DB_HOST=%s
DB_PORT=3306
DB_NAME=%s
DB_DATABASE=%s
DB_USER=ace_user
DB_USERNAME=ace_user
DB_PASSWORD=secret
MYSQL_ROOT_PASSWORD=rootsecret
MYSQL_DATABASE=%s
MYSQL_USER=ace_user
MYSQL_PASSWORD=secret
`, host, cfg.DBName(), cfg.DBName(), cfg.DBName())

	if err := appendToEnv(base, envConfig); err != nil {
		return err
	}

	switch cfg.Framework.ID {
	case "django":
		reqPath := filepath.Join(base, "requirements.txt")
		if reqBytes, err := os.ReadFile(reqPath); err == nil {
			reqContent := string(reqBytes)
			if !strings.Contains(reqContent, "pymysql") {
				reqContent += "pymysql>=1.1.0\n"
				_ = os.WriteFile(reqPath, []byte(reqContent), 0644)
			}
		}
		pkgName := strings.ReplaceAll(strings.ToLower(cfg.Name), "-", "_")
		initPath := filepath.Join(base, pkgName, "__init__.py")
		if initBytes, err := os.ReadFile(initPath); err == nil {
			initStr := string(initBytes)
			if !strings.Contains(initStr, "pymysql") {
				initStr += "\nimport pymysql\npymysql.install_as_MySQLdb()\n"
				_ = os.WriteFile(initPath, []byte(initStr), 0644)
			}
		}

	case "fastapi":
		reqPath := filepath.Join(base, "requirements.txt")
		if reqBytes, err := os.ReadFile(reqPath); err == nil {
			reqContent := string(reqBytes)
			if !strings.Contains(reqContent, "aiomysql") {
				reqContent += "sqlalchemy>=2.0.30\naiomysql>=0.2.0\ncryptography>=42.0.0\n"
				_ = os.WriteFile(reqPath, []byte(reqContent), 0644)
			}
		}
		sessionPy := fmt.Sprintf(`from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from sqlalchemy.orm import declarative_base
import os

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "mysql+aiomysql://ace_user:secret@%s:3306/%s",
)

engine = create_async_engine(DATABASE_URL, echo=True)
AsyncSessionLocal = async_sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

Base = declarative_base()

async def get_db():
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()
`, host, cfg.DBName())
		_ = os.MkdirAll(filepath.Join(base, "app", "db"), 0755)
		_ = os.WriteFile(filepath.Join(base, "app/db/session.py"), []byte(sessionPy), 0644)

	case "fiber", "gin":
		goModPath := filepath.Join(base, "go.mod")
		if modBytes, err := os.ReadFile(goModPath); err == nil {
			modStr := string(modBytes)
			if !strings.Contains(modStr, "gorm.io/driver/mysql") {
				modStr = strings.Replace(
					modStr,
					"require (",
					"require (\n\tgorm.io/gorm v1.25.10\n\tgorm.io/driver/mysql v1.5.7",
					1,
				)
				_ = os.WriteFile(goModPath, []byte(modStr), 0644)
			}
		}
		dbGo := fmt.Sprintf(`package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "%s"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "ace_user"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "secret"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "%s"
	}

	dsn := fmt.Sprintf("%%s:%%s@tcp(%%s:%%s)/%%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Failed to connect to MySQL: %%v", err)
		return
	}

	log.Println("Connected to MySQL database successfully!")
}
`, host, cfg.DBName())
		_ = os.MkdirAll(filepath.Join(base, "internal", "database"), 0755)
		_ = os.WriteFile(filepath.Join(base, "internal/database/database.go"), []byte(dbGo), 0644)

	case "next":
		prismaPath := filepath.Join(base, "prisma", "schema.prisma")
		if pBytes, err := os.ReadFile(prismaPath); err == nil {
			pStr := strings.Replace(string(pBytes), `provider = "postgresql"`, `provider = "mysql"`, 1)
			_ = os.WriteFile(prismaPath, []byte(pStr), 0644)
		}

	case "express", "nestjs":
		pkgPath := filepath.Join(base, "package.json")
		if pkgBytes, err := os.ReadFile(pkgPath); err == nil {
			pkgStr := string(pkgBytes)
			if !strings.Contains(pkgStr, `"mysql2"`) {
				pkgStr = strings.Replace(
					pkgStr,
					`"dependencies": {`,
					"\"dependencies\": {\n    \"mysql2\": \"^3.9.8\",",
					1,
				)
				_ = os.WriteFile(pkgPath, []byte(pkgStr), 0644)
			}
		}
	}

	return nil
}

// InjectSQLite configures SQLite database settings and drivers
func InjectSQLite(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	envConfig := `
# SQLite Database Configuration
DB_CONNECTION=sqlite
DB_ENGINE=django.db.backends.sqlite3
DB_NAME=db.sqlite3
DB_DATABASE=database/database.sqlite
`
	if err := appendToEnv(base, envConfig); err != nil {
		return err
	}

	switch cfg.Framework.ID {
	case "fastapi":
		reqPath := filepath.Join(base, "requirements.txt")
		if reqBytes, err := os.ReadFile(reqPath); err == nil {
			reqContent := string(reqBytes)
			if !strings.Contains(reqContent, "aiosqlite") {
				reqContent += "sqlalchemy>=2.0.30\naiosqlite>=0.20.0\n"
				_ = os.WriteFile(reqPath, []byte(reqContent), 0644)
			}
		}
		sessionPy := `from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from sqlalchemy.orm import declarative_base
import os

DATABASE_URL = os.getenv("DATABASE_URL", "sqlite+aiosqlite:///./app.db")

engine = create_async_engine(DATABASE_URL, echo=True)
AsyncSessionLocal = async_sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

Base = declarative_base()

async def get_db():
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()
`
		_ = os.MkdirAll(filepath.Join(base, "app", "db"), 0755)
		_ = os.WriteFile(filepath.Join(base, "app/db/session.py"), []byte(sessionPy), 0644)

	case "fiber", "gin":
		goModPath := filepath.Join(base, "go.mod")
		if modBytes, err := os.ReadFile(goModPath); err == nil {
			modStr := string(modBytes)
			if !strings.Contains(modStr, "gorm.io/driver/sqlite") {
				modStr = strings.Replace(
					modStr,
					"require (",
					"require (\n\tgorm.io/gorm v1.25.10\n\tgorm.io/driver/sqlite v1.5.6",
					1,
				)
				_ = os.WriteFile(goModPath, []byte(modStr), 0644)
			}
		}

	case "next":
		prismaPath := filepath.Join(base, "prisma", "schema.prisma")
		if pBytes, err := os.ReadFile(prismaPath); err == nil {
			pStr := strings.Replace(string(pBytes), `provider = "postgresql"`, `provider = "sqlite"`, 1)
			_ = os.WriteFile(prismaPath, []byte(pStr), 0644)
		}
	}

	return nil
}

// InjectMongo configures MongoDB connection settings and drivers
func InjectMongo(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath
	host := "127.0.0.1"
	if cfg.Docker {
		host = "db"
	}

	envConfig := fmt.Sprintf(`
# MongoDB Configuration
MONGODB_URI=mongodb://ace_user:secret@%s:27017/%s?authSource=admin
MONGO_INITDB_ROOT_USERNAME=ace_user
MONGO_INITDB_ROOT_PASSWORD=secret
MONGO_INITDB_DATABASE=%s
`, host, cfg.DBName(), cfg.DBName())

	if err := appendToEnv(base, envConfig); err != nil {
		return err
	}

	switch cfg.Framework.ID {
	case "fastapi":
		reqPath := filepath.Join(base, "requirements.txt")
		if reqBytes, err := os.ReadFile(reqPath); err == nil {
			reqContent := string(reqBytes)
			if !strings.Contains(reqContent, "motor") {
				reqContent += "motor>=3.4.0\n"
				_ = os.WriteFile(reqPath, []byte(reqContent), 0644)
			}
		}
	case "fiber", "gin":
		goModPath := filepath.Join(base, "go.mod")
		if modBytes, err := os.ReadFile(goModPath); err == nil {
			modStr := string(modBytes)
			if !strings.Contains(modStr, "go.mongodb.org/mongo-driver") {
				modStr = strings.Replace(
					modStr,
					"require (",
					"require (\n\tgo.mongodb.org/mongo-driver/mongo v1.15.0",
					1,
				)
				_ = os.WriteFile(goModPath, []byte(modStr), 0644)
			}
		}
	case "express", "nestjs":
		pkgPath := filepath.Join(base, "package.json")
		if pkgBytes, err := os.ReadFile(pkgPath); err == nil {
			pkgStr := string(pkgBytes)
			if !strings.Contains(pkgStr, `"mongoose"`) {
				pkgStr = strings.Replace(
					pkgStr,
					`"dependencies": {`,
					"\"dependencies\": {\n    \"mongoose\": \"^8.4.1\",",
					1,
				)
				_ = os.WriteFile(pkgPath, []byte(pkgStr), 0644)
			}
		}
	}

	return nil
}

// InjectRedis configures Redis standalone database settings
func InjectRedis(cfg *config.ProjectConfig) error {
	return InjectCompanionRedis(cfg)
}
