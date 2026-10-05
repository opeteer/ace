package framework

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldFastAPI generates a modern, production-grade async FastAPI project structure
func ScaffoldFastAPI(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"app/api/v1/endpoints",
		"app/core",
		"app/db",
		"app/models",
		"app/schemas",
		"tests",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"requirements.txt": `fastapi>=0.111.0
uvicorn[standard]>=0.30.0
pydantic>=2.7.0
pydantic-settings>=2.2.0
python-dotenv>=1.0.0
`,

		"pyproject.toml": fmt.Sprintf(`[project]
name = "%s"
version = "0.1.0"
description = "Scaffolded with Ace CLI"
readme = "README.md"
requires-python = ">=3.11"
dependencies = [
    "fastapi>=0.111.0",
    "uvicorn[standard]>=0.30.0",
    "pydantic>=2.7.0",
    "pydantic-settings>=2.2.0",
    "python-dotenv>=1.0.0",
]

[tool.pytest.ini_options]
asyncio_mode = "auto"
testpaths = ["tests"]
`, cfg.Name),

		"app/__init__.py": `"""Application package."""
`,

		"app/main.py": fmt.Sprintf(`from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.core.config import settings
from app.api.v1.api import api_router

app = FastAPI(
    title="%s",
    openapi_url=f"{settings.API_V1_STR}/openapi.json",
    docs_url="/docs",
    redoc_url="/redoc",
)

if settings.BACKEND_CORS_ORIGINS:
    app.add_middleware(
        CORSMiddleware,
        allow_origins=[str(origin) for origin in settings.BACKEND_CORS_ORIGINS],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

app.include_router(api_router, prefix=settings.API_V1_STR)

@app.get("/")
async def root():
    return {
        "app": "%s",
        "status": "online",
        "docs": "/docs",
        "scaffolded_by": "Ace CLI",
    }
`, cfg.Name, cfg.Name),

		"app/core/__init__.py": ``,

		"app/core/config.py": fmt.Sprintf(`from typing import List
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    PROJECT_NAME: str = "%s"
    API_V1_STR: str = "/api/v1"
    BACKEND_CORS_ORIGINS: List[str] = ["*"]

    class Config:
        case_sensitive = True
        env_file = ".env"

settings = Settings()
`, cfg.Name),

		"app/api/__init__.py": ``,
		"app/api/v1/__init__.py": ``,

		"app/api/v1/api.py": `from fastapi import APIRouter
from app.api.v1.endpoints import health

api_router = APIRouter()
api_router.include_router(health.router, prefix="/health", tags=["health"])
`,

		"app/api/v1/endpoints/__init__.py": ``,

		"app/api/v1/endpoints/health.py": `from fastapi import APIRouter
from datetime import datetime, timezone

router = APIRouter()

@router.get("")
async def health_check():
    return {
        "status": "healthy",
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }
`,

		"tests/__init__.py": ``,

		"tests/test_health.py": `from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_root():
    response = client.get("/")
    assert response.status_code == 200
    assert response.json()["status"] == "online"

def test_health():
    response = client.get("/api/v1/health")
    assert response.status_code == 200
    assert response.json()["status"] == "healthy"
`,

		".gitignore": `__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
.env
.venv
env/
venv/
ENV/
.pytest_cache/
.coverage
htmlcov/
`,

		".env.example": fmt.Sprintf(`PROJECT_NAME=%s
API_V1_STR=/api/v1
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
