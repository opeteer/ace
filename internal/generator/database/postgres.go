package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// InjectPostgres wires PostgreSQL into the target project according to its framework
func InjectPostgres(cfg *config.ProjectConfig) error {
	host := "127.0.0.1"
	if cfg.Docker {
		host = "db"
	}

	switch cfg.Framework.ID {
	case "laravel":
		return injectLaravelPostgres(cfg, host)
	case "fastapi":
		return injectFastAPIPostgres(cfg, host)
	case "django":
		return injectDjangoPostgres(cfg, host)
	case "next":
		return injectNextJSPostgres(cfg, host)
	case "fiber", "gin":
		return injectFiberPostgres(cfg, host)
	default:
		return injectGenericDB(cfg)
	}
}

func injectDjangoPostgres(cfg *config.ProjectConfig, host string) error {
	envConfig := fmt.Sprintf(`
# Database Configuration (PostgreSQL)
DB_ENGINE=django.db.backends.postgresql
DB_NAME=%s
DB_USER=ace_user
DB_PASSWORD=secret
DB_HOST=%s
DB_PORT=5432
POSTGRES_USER=ace_user
POSTGRES_PASSWORD=secret
POSTGRES_DB=%s
`, cfg.DBName(), host, cfg.DBName())

	return appendToEnv(cfg.TargetPath, envConfig)
}

func injectLaravelPostgres(cfg *config.ProjectConfig, host string) error {
	envConfig := fmt.Sprintf(`
# PostgreSQL Database Settings
DB_CONNECTION=pgsql
DB_HOST=%s
DB_PORT=5432
DB_DATABASE=%s
DB_USERNAME=ace_user
DB_PASSWORD=secret
`, host, cfg.DBName())

	return appendToEnv(cfg.TargetPath, envConfig)
}

func injectFastAPIPostgres(cfg *config.ProjectConfig, host string) error {
	base := cfg.TargetPath

	sessionPy := fmt.Sprintf(`from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from sqlalchemy.orm import declarative_base
import os

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgresql+asyncpg://ace_user:secret@%s:5432/%s",
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

	itemPy := `from sqlalchemy import Column, Integer, String, DateTime, func
from app.db.session import Base

class Item(Base):
    __tablename__ = "items"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(255), nullable=False)
    description = Column(String(1000), nullable=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
`

	if err := os.WriteFile(filepath.Join(base, "app/db/session.py"), []byte(sessionPy), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "app/models/item.py"), []byte(itemPy), 0644); err != nil {
		return err
	}

	// Append dependencies to requirements.txt
	reqPath := filepath.Join(base, "requirements.txt")
	if reqBytes, err := os.ReadFile(reqPath); err == nil {
		reqContent := string(reqBytes) + "sqlalchemy>=2.0.30\nasyncpg>=0.29.0\nalembic>=1.13.1\n"
		_ = os.WriteFile(reqPath, []byte(reqContent), 0644)
	}

	envConfig := fmt.Sprintf(`
# Database Configuration (PostgreSQL)
DATABASE_URL=postgresql+asyncpg://ace_user:secret@%s:5432/%s
POSTGRES_USER=ace_user
POSTGRES_PASSWORD=secret
POSTGRES_DB=%s
`, host, cfg.DBName(), cfg.DBName())

	return appendToEnv(base, envConfig)
}

func injectNextJSPostgres(cfg *config.ProjectConfig, host string) error {
	base := cfg.TargetPath

	if err := os.MkdirAll(filepath.Join(base, "prisma"), 0755); err != nil {
		return err
	}

	prismaSchema := fmt.Sprintf(`datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

model User {
  id        String   @id @default(uuid())
  email     String   @unique
  name      String?
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
}
`)

	prismaLib := `import { PrismaClient } from "@prisma/client";

const globalForPrisma = globalThis as unknown as {
  prisma: PrismaClient | undefined;
};

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    log: process.env.NODE_ENV === "development" ? ["query", "error", "warn"] : ["error"],
  });

if (process.env.NODE_ENV !== "production") globalForPrisma.prisma = prisma;
`

	if err := os.WriteFile(filepath.Join(base, "prisma/schema.prisma"), []byte(prismaSchema), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "lib/prisma.ts"), []byte(prismaLib), 0644); err != nil {
		return err
	}

	// Update package.json to include prisma
	pkgPath := filepath.Join(base, "package.json")
	if pkgBytes, err := os.ReadFile(pkgPath); err == nil {
		pkgStr := string(pkgBytes)
		if !strings.Contains(pkgStr, "@prisma/client") {
			pkgStr = strings.Replace(
				pkgStr,
				`"dependencies": {`,
				"\"dependencies\": {\n    \"@prisma/client\": \"^5.16.1\",",
				1,
			)
			pkgStr = strings.Replace(
				pkgStr,
				`"devDependencies": {`,
				"\"devDependencies\": {\n    \"prisma\": \"^5.16.1\",",
				1,
			)
			_ = os.WriteFile(pkgPath, []byte(pkgStr), 0644)
		}
	}

	envConfig := fmt.Sprintf(`
# Database Configuration (PostgreSQL with Prisma)
DATABASE_URL=postgresql://ace_user:secret@%s:5432/%s?schema=public
POSTGRES_USER=ace_user
POSTGRES_PASSWORD=secret
POSTGRES_DB=%s
`, host, cfg.DBName(), cfg.DBName())

	return appendToEnv(base, envConfig)
}

func injectFiberPostgres(cfg *config.ProjectConfig, host string) error {
	base := cfg.TargetPath

	dbGo := fmt.Sprintf(`package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
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
		port = "5432"
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

	dsn := fmt.Sprintf("host=%%s user=%%s password=%%s dbname=%%s port=%%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %%v", err)
		return
	}

	log.Println("Connected to PostgreSQL database successfully!")
}
`, host, cfg.DBName())

	_ = os.MkdirAll(filepath.Join(base, "internal", "database"), 0755)
	if err := os.WriteFile(filepath.Join(base, "internal/database/database.go"), []byte(dbGo), 0644); err != nil {
		return err
	}

	// Update go.mod to include gorm
	goModPath := filepath.Join(base, "go.mod")
	if modBytes, err := os.ReadFile(goModPath); err == nil {
		modStr := string(modBytes)
		if !strings.Contains(modStr, "gorm.io/gorm") {
			modStr = strings.Replace(
				modStr,
				`require (`,
				"require (\n\tgorm.io/gorm v1.25.10\n\tgorm.io/driver/postgres v1.5.9",
				1,
			)
			_ = os.WriteFile(goModPath, []byte(modStr), 0644)
		}
	}

	envConfig := fmt.Sprintf(`
# Database Configuration (PostgreSQL)
DB_CONNECTION=postgres
DB_HOST=%s
DB_PORT=5432
DB_USER=ace_user
DB_PASSWORD=secret
DB_NAME=%s
`, host, cfg.DBName())

	return appendToEnv(base, envConfig)
}
