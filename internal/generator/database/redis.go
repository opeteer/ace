package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// InjectCompanionRedis configures companion Redis cache and queue support across all frameworks
func InjectCompanionRedis(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	host := "127.0.0.1"
	if cfg.Docker {
		host = "redis"
	}

	envConfig := fmt.Sprintf(`
# Redis In-Memory Cache & Message Broker
REDIS_HOST=%s
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_URL=redis://%s:6379/0
`, host, host)

	if err := appendToEnv(base, envConfig); err != nil {
		return err
	}

	switch cfg.Framework.ID {
	case "django":
		return injectDjangoRedis(cfg)
	case "fastapi":
		return injectFastAPIRedis(cfg, host)
	case "fiber", "gin":
		return injectGoRedis(cfg)
	case "laravel":
		return injectLaravelRedis(cfg)
	case "springboot":
		return injectSpringBootRedis(cfg, host)
	case "aspnet":
		return injectAspNetRedis(cfg)
	case "axum":
		return injectAxumRedis(cfg)
	default:
		// Node.js stacks: next, nuxt, sveltekit, astro, express, nestjs, vite-react, vite-vue, expo
		return injectNodeRedis(cfg)
	}
}

func injectDjangoRedis(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	// Append dependencies to requirements.txt
	reqPath := filepath.Join(base, "requirements.txt")
	if reqBytes, err := os.ReadFile(reqPath); err == nil {
		reqContent := string(reqBytes)
		if !strings.Contains(reqContent, "django-redis") {
			reqContent += "django-redis>=5.4.0\nredis>=5.0.0\n"
			_ = os.WriteFile(reqPath, []byte(reqContent), 0644)
		}
	}

	// Append cache setting to settings.py
	pkgName := strings.ReplaceAll(strings.ToLower(cfg.Name), "-", "_")
	settingsPath := filepath.Join(base, pkgName, "settings.py")
	if setBytes, err := os.ReadFile(settingsPath); err == nil {
		setStr := string(setBytes)
		if !strings.Contains(setStr, "django_redis") {
			cacheConfig := `
# Redis Cache Configuration
CACHES = {
    "default": {
        "BACKEND": "django_redis.cache.RedisCache",
        "LOCATION": os.getenv("REDIS_URL", "redis://127.0.0.1:6379/1"),
        "OPTIONS": {
            "CLIENT_CLASS": "django_redis.client.DefaultClient",
        }
    }
}
`
			setStr += cacheConfig
			_ = os.WriteFile(settingsPath, []byte(setStr), 0644)
		}
	}

	return nil
}

func injectFastAPIRedis(cfg *config.ProjectConfig, host string) error {
	base := cfg.TargetPath

	reqPath := filepath.Join(base, "requirements.txt")
	if reqBytes, err := os.ReadFile(reqPath); err == nil {
		reqContent := string(reqBytes)
		if !strings.Contains(reqContent, "redis") {
			reqContent += "redis>=5.0.0\n"
			_ = os.WriteFile(reqPath, []byte(reqContent), 0644)
		}
	}

	redisPy := fmt.Sprintf(`import os
from redis import asyncio as aioredis

REDIS_URL = os.getenv("REDIS_URL", "redis://%s:6379/0")

async def get_redis_client():
    client = aioredis.from_url(REDIS_URL, decode_responses=True)
    try:
        yield client
    finally:
        await client.aclose()
`, host)

	_ = os.MkdirAll(filepath.Join(base, "app", "core"), 0755)
	return os.WriteFile(filepath.Join(base, "app", "core", "redis.py"), []byte(redisPy), 0644)
}

func injectGoRedis(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	goModPath := filepath.Join(base, "go.mod")
	if modBytes, err := os.ReadFile(goModPath); err == nil {
		modStr := string(modBytes)
		if !strings.Contains(modStr, "github.com/redis/go-redis") {
			modStr = strings.Replace(
				modStr,
				"require (",
				"require (\n\tgithub.com/redis/go-redis/v9 v9.5.3",
				1,
			)
			_ = os.WriteFile(goModPath, []byte(modStr), 0644)
		}
	}

	redisGo := `package database

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	pass := os.Getenv("REDIS_PASSWORD")

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: pass,
		DB:       0,
	})
	return RDB
}

func PingRedis(ctx context.Context) error {
	if RDB == nil {
		InitRedis()
	}
	return RDB.Ping(ctx).Err()
}
`

	_ = os.MkdirAll(filepath.Join(base, "internal", "database"), 0755)
	return os.WriteFile(filepath.Join(base, "internal", "database", "redis.go"), []byte(redisGo), 0644)
}

func injectLaravelRedis(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	composerPath := filepath.Join(base, "composer.json")
	if compBytes, err := os.ReadFile(composerPath); err == nil {
		compStr := string(compBytes)
		if !strings.Contains(compStr, "predis/predis") {
			compStr = strings.Replace(
				compStr,
				`"require": {`,
				"\"require\": {\n        \"predis/predis\": \"^2.2\",",
				1,
			)
			_ = os.WriteFile(composerPath, []byte(compStr), 0644)
		}
	}

	laravelEnv := `REDIS_CLIENT=predis
CACHE_STORE=redis
QUEUE_CONNECTION=redis
`
	return appendToEnv(base, laravelEnv)
}

func injectSpringBootRedis(cfg *config.ProjectConfig, host string) error {
	base := cfg.TargetPath

	pomPath := filepath.Join(base, "pom.xml")
	if pomBytes, err := os.ReadFile(pomPath); err == nil {
		pomStr := string(pomBytes)
		if !strings.Contains(pomStr, "spring-boot-starter-data-redis") {
			pomStr = strings.Replace(
				pomStr,
				"<dependencies>",
				`<dependencies>
		<dependency>
			<groupId>org.springframework.boot</groupId>
			<artifactId>spring-boot-starter-data-redis</artifactId>
		</dependency>`,
				1,
			)
			_ = os.WriteFile(pomPath, []byte(pomStr), 0644)
		}
	}

	propsPath := filepath.Join(base, "src", "main", "resources", "application.properties")
	if propBytes, err := os.ReadFile(propsPath); err == nil {
		propStr := string(propBytes)
		if !strings.Contains(propStr, "spring.data.redis") {
			propStr += fmt.Sprintf("\nspring.data.redis.host=${REDIS_HOST:%s}\nspring.data.redis.port=${REDIS_PORT:6379}\n", host)
			_ = os.WriteFile(propsPath, []byte(propStr), 0644)
		}
	}

	return nil
}

func injectAspNetRedis(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	matches, _ := filepath.Glob(filepath.Join(base, "*.csproj"))
	for _, csproj := range matches {
		csBytes, err := os.ReadFile(csproj)
		if err == nil {
			csStr := string(csBytes)
			if !strings.Contains(csStr, "StackExchangeRedis") {
				csStr = strings.Replace(
					csStr,
					"</ItemGroup>",
					"  <PackageReference Include=\"Microsoft.Extensions.Caching.StackExchangeRedis\" Version=\"8.0.4\" />\n  </ItemGroup>",
					1,
				)
				_ = os.WriteFile(csproj, []byte(csStr), 0644)
			}
		}
	}
	return nil
}

func injectAxumRedis(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	cargoPath := filepath.Join(base, "Cargo.toml")
	if cargoBytes, err := os.ReadFile(cargoPath); err == nil {
		cargoStr := string(cargoBytes)
		if !strings.Contains(cargoStr, "redis =") {
			cargoStr += "\nredis = { version = \"0.25\", features = [\"tokio-comp\"] }\n"
			_ = os.WriteFile(cargoPath, []byte(cargoStr), 0644)
		}
	}
	return nil
}

func injectNodeRedis(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	pkgPath := filepath.Join(base, "package.json")
	if pkgBytes, err := os.ReadFile(pkgPath); err == nil {
		pkgStr := string(pkgBytes)
		if !strings.Contains(pkgStr, `"ioredis"`) {
			if !strings.Contains(pkgStr, `"dependencies": {`) {
				if strings.Contains(pkgStr, `"devDependencies": {`) {
					pkgStr = strings.Replace(pkgStr, `"devDependencies": {`, "\"dependencies\": {\n  },\n  \"devDependencies\": {", 1)
				} else {
					last := strings.LastIndex(pkgStr, "}")
					if last != -1 {
						pkgStr = pkgStr[:last] + ",\n  \"dependencies\": {\n  }\n}"
					}
				}
			}
			if strings.Contains(pkgStr, `"dependencies": {}`) {
				pkgStr = strings.Replace(pkgStr, `"dependencies": {}`, "\"dependencies\": {\n    \"ioredis\": \"^5.4.1\"\n  }", 1)
			} else if strings.Contains(pkgStr, "\"dependencies\": {\n  }") {
				pkgStr = strings.Replace(pkgStr, "\"dependencies\": {\n  }", "\"dependencies\": {\n    \"ioredis\": \"^5.4.1\"\n  }", 1)
			} else {
				pkgStr = strings.Replace(
					pkgStr,
					`"dependencies": {`,
					"\"dependencies\": {\n    \"ioredis\": \"^5.4.1\",",
					1,
				)
			}
			_ = os.WriteFile(pkgPath, []byte(pkgStr), 0644)
		}
	}

	redisTs := `import Redis from 'ioredis';

const redisHost = process.env.REDIS_HOST || '127.0.0.1';
const redisPort = parseInt(process.env.REDIS_PORT || '6379', 10);
const redisPassword = process.env.REDIS_PASSWORD || undefined;

export const redis = new Redis({
  host: redisHost,
  port: redisPort,
  password: redisPassword,
  lazyConnect: true,
});
`

	// Write to src/lib/redis.ts or lib/redis.ts depending on folder structure
	targetDir := filepath.Join(base, "src", "lib")
	if _, err := os.Stat(filepath.Join(base, "src")); os.IsNotExist(err) {
		targetDir = filepath.Join(base, "lib")
	}
	_ = os.MkdirAll(targetDir, 0755)
	return os.WriteFile(filepath.Join(targetDir, "redis.ts"), []byte(redisTs), 0644)
}
