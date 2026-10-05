package config

import "strings"

// SupportedFrameworks contains the full catalog of frameworks supported by Ace
var SupportedFrameworks = []FrameworkSpec{
	// Fullstack / SSR
	{ID: "next", Name: "Next.js (App Router)", Category: CategoryFullstack, Language: "TypeScript", DefaultPort: 3000, Description: "Modern React fullstack framework with server components"},
	{ID: "nuxt", Name: "Nuxt 3", Category: CategoryFullstack, Language: "TypeScript", DefaultPort: 3000, Description: "Intuitive Vue fullstack framework"},
	{ID: "sveltekit", Name: "SvelteKit", Category: CategoryFullstack, Language: "TypeScript", DefaultPort: 5173, Description: "Fast, minimal bundle Svelte fullstack framework"},
	{ID: "astro", Name: "Astro", Category: CategoryFullstack, Language: "TypeScript", DefaultPort: 4321, Description: "Content-driven web framework with islands architecture"},
	{ID: "laravel", Name: "Laravel 11", Category: CategoryFullstack, Language: "PHP", DefaultPort: 8000, Description: "The PHP framework for web artisans with rich ecosystem"},
	{ID: "django", Name: "Django 5", Category: CategoryFullstack, Language: "Python", DefaultPort: 8000, Description: "Batteries-included high-level Python web framework"},
	{ID: "rails", Name: "Ruby on Rails 7", Category: CategoryFullstack, Language: "Ruby", DefaultPort: 3000, Description: "Convention over configuration fullstack framework"},

	// Backend APIs
	{ID: "fastapi", Name: "FastAPI", Category: CategoryBackend, Language: "Python", DefaultPort: 8000, Description: "High-performance async Python API with automatic OpenAPI docs"},
	{ID: "express", Name: "Express.js", Category: CategoryBackend, Language: "TypeScript", DefaultPort: 4000, Description: "Fast, unopinionated, minimalist web framework for Node.js"},
	{ID: "nestjs", Name: "NestJS", Category: CategoryBackend, Language: "TypeScript", DefaultPort: 3000, Description: "Progressive enterprise TypeScript framework with modular DI"},
	{ID: "fiber", Name: "Go Fiber", Category: CategoryBackend, Language: "Go", DefaultPort: 8080, Description: "Express-inspired ultra-fast web framework built on Fasthttp"},
	{ID: "gin", Name: "Go Gin", Category: CategoryBackend, Language: "Go", DefaultPort: 8080, Description: "Battle-tested, lightweight HTTP web framework for Go"},
	{ID: "axum", Name: "Axum", Category: CategoryBackend, Language: "Rust", DefaultPort: 3000, Description: "Ergonomic and modular async web framework for Rust"},
	{ID: "springboot", Name: "Spring Boot 3", Category: CategoryBackend, Language: "Java/Kotlin", DefaultPort: 8080, Description: "Production-ready enterprise framework for the JVM"},
	{ID: "aspnet", Name: "ASP.NET Core Web API", Category: CategoryBackend, Language: "C#", DefaultPort: 5000, Description: "Cross-platform, high-performance Microsoft web API stack"},

	// Frontend SPA
	{ID: "vite-react", Name: "React (Vite)", Category: CategoryFrontend, Language: "TypeScript", DefaultPort: 5173, Description: "Client-side React application powered by Vite"},
	{ID: "vite-vue", Name: "Vue 3 (Vite)", Category: CategoryFrontend, Language: "TypeScript", DefaultPort: 5173, Description: "Client-side Vue 3 application powered by Vite"},

	// Mobile (Cross-Platform)
	{ID: "flutter", Name: "Flutter", Category: CategoryMobile, Language: "Dart", DefaultPort: 0, Description: "Google's UI toolkit for natively compiled mobile, web, and desktop"},
	{ID: "expo", Name: "React Native (Expo)", Category: CategoryMobile, Language: "TypeScript", DefaultPort: 8081, Description: "Universal React Native framework for Android and iOS"},
}

// SupportedDatabases contains the full catalog of database engines supported by Ace
var SupportedDatabases = []DatabaseSpec{
	{ID: "postgres", Name: "PostgreSQL 16+", Paradigm: "Relational (SQL)", DefaultPort: 5432, DockerImage: "postgres:16-alpine", EnvPrefix: "POSTGRES"},
	{ID: "mysql", Name: "MySQL 8.4 LTS", Paradigm: "Relational (SQL)", DefaultPort: 3306, DockerImage: "mysql:8.4", EnvPrefix: "MYSQL"},
	{ID: "mariadb", Name: "MariaDB 11", Paradigm: "Relational (SQL)", DefaultPort: 3306, DockerImage: "mariadb:11", EnvPrefix: "MARIADB"},
	{ID: "sqlite", Name: "SQLite 3", Paradigm: "Embedded (SQL)", DefaultPort: 0, DockerImage: "", EnvPrefix: "SQLITE"},
	{ID: "mongo", Name: "MongoDB 7", Paradigm: "Document (NoSQL)", DefaultPort: 27017, DockerImage: "mongo:7-jammy", EnvPrefix: "MONGO"},
	{ID: "redis", Name: "Redis 7 / Valkey", Paradigm: "Key-Value / In-Memory", DefaultPort: 6379, DockerImage: "redis:7-alpine", EnvPrefix: "REDIS"},
	{ID: "clickhouse", Name: "ClickHouse", Paradigm: "Analytical (OLAP)", DefaultPort: 8123, DockerImage: "clickhouse/clickhouse-server:latest", EnvPrefix: "CLICKHOUSE"},
	{ID: "meilisearch", Name: "Meilisearch", Paradigm: "Search Engine", DefaultPort: 7700, DockerImage: "getmeili/meilisearch:v1.8", EnvPrefix: "MEILI"},
	{ID: "elastic", Name: "Elasticsearch 8", Paradigm: "Search Engine", DefaultPort: 9200, DockerImage: "elasticsearch:8.13.0", EnvPrefix: "ELASTIC"},
	{ID: "neo4j", Name: "Neo4j Community", Paradigm: "Graph Database", DefaultPort: 7474, DockerImage: "neo4j:5-community", EnvPrefix: "NEO4J"},
}

// FindFramework searches for a framework by ID or alias (case-insensitive)
func FindFramework(query string) (FrameworkSpec, bool) {
	q := strings.ToLower(strings.TrimSpace(query))
	switch q {
	case "next", "nextjs", "next.js":
		q = "next"
	case "nuxt", "nuxtjs":
		q = "nuxt"
	case "svelte", "sveltekit":
		q = "sveltekit"
	case "react", "vite-react":
		q = "vite-react"
	case "vue", "vite-vue":
		q = "vite-vue"
	case "spring", "springboot", "spring-boot":
		q = "springboot"
	case "dotnet", "csharp", "aspnet", "aspnetcore":
		q = "aspnet"
	case "rn", "reactnative", "react-native", "expo":
		q = "expo"
	}

	for _, f := range SupportedFrameworks {
		if strings.EqualFold(f.ID, q) {
			return f, true
		}
	}
	return FrameworkSpec{}, false
}

// FindDatabase searches for a database by ID or alias (case-insensitive)
func FindDatabase(query string) (DatabaseSpec, bool) {
	q := strings.ToLower(strings.TrimSpace(query))
	switch q {
	case "postgres", "postgresql", "pgsql", "pg":
		q = "postgres"
	case "mariadb", "maria":
		q = "mariadb"
	case "mongodb", "mongo":
		q = "mongo"
	case "valkey", "redis":
		q = "redis"
	case "elasticsearch", "elastic":
		q = "elastic"
	case "meili", "meilisearch":
		q = "meilisearch"
	}

	for _, d := range SupportedDatabases {
		if strings.EqualFold(d.ID, q) {
			return d, true
		}
	}
	return DatabaseSpec{}, false
}
