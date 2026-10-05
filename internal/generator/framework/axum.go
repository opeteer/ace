package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// ScaffoldAxum generates a high-performance async Rust Axum microservice
func ScaffoldAxum(cfg *config.ProjectConfig) error {
	base := cfg.TargetPath

	dirs := []string{
		"src",
		"tests",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0755); err != nil {
			return err
		}
	}

	crateName := strings.ReplaceAll(strings.ToLower(cfg.Name), "-", "_")

	files := map[string]string{
		"Cargo.toml": fmt.Sprintf(`[package]
name = "%s"
version = "0.1.0"
edition = "2021"

[dependencies]
axum = "0.7.5"
tokio = { version = "1.38", features = ["full"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tower-http = { version = "0.5.2", features = ["cors", "trace"] }
tracing = "0.1"
tracing-subscriber = "0.3"
dotenvy = "0.15"
`, crateName),

		"src/main.rs": fmt.Sprintf(`use axum::{routing::get, Json, Router};
use serde::Serialize;
use std::net::SocketAddr;
use tower_http::cors::CorsLayer;

#[derive(Serialize)]
struct RootResponse {
    app: &'static str,
    framework: &'static str,
    status: &'static str,
    scaffolded_by: &'static str,
}

#[derive(Serialize)]
struct HealthResponse {
    status: &'static str,
    timestamp: String,
}

async fn root() -> Json<RootResponse> {
    Json(RootResponse {
        app: "%s",
        framework: "Rust Axum",
        status: "online",
        scaffolded_by: "Ace CLI",
    })
}

async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy",
        timestamp: chrono_now(),
    })
}

fn chrono_now() -> String {
    // Simple ISO timestamp format
    let duration = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap_or_default();
    format!("{}s-since-epoch", duration.as_secs())
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    dotenvy::dotenv().ok();

    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "%d".to_string())
        .parse()
        .unwrap_or(%d);

    let app = Router::new()
        .route("/", get(root))
        .route("/api/health", get(health))
        .layer(CorsLayer::permissive());

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!(">> Axum server running on http://{}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
`, cfg.Name, cfg.Framework.DefaultPort, cfg.Framework.DefaultPort),

		".env.example": fmt.Sprintf(`PORT=%d
RUST_LOG=info
`, cfg.Framework.DefaultPort),

		".gitignore": `/target
**/*.rs.bk
Cargo.lock
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
