package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// driverSpec describes the client library each ecosystem needs for a database.
type driverSpec struct {
	nodeDeps   map[string]string // package.json dependencies
	goRequires []string          // go.mod require lines
	mavenDeps  [][2]string       // groupId, artifactId (version managed by Spring Boot BOM)
	nugetPkgs  [][2]string       // package, version
	cargoLines []string          // Cargo.toml dependency lines
	gems       []string          // Gemfile lines
	springURL  string            // spring.datasource.url template (%s host, %s db)
	springDrv  string
}

var driverSpecs = map[string]driverSpec{
	"postgres": {
		nodeDeps:   map[string]string{"pg": "^8.11.5"},
		goRequires: []string{"gorm.io/gorm v1.25.10", "gorm.io/driver/postgres v1.5.9"},
		mavenDeps:  [][2]string{{"org.postgresql", "postgresql"}},
		nugetPkgs:  [][2]string{{"Npgsql.EntityFrameworkCore.PostgreSQL", "8.0.4"}},
		cargoLines: []string{`sqlx = { version = "0.7", features = ["runtime-tokio", "postgres"] }`},
		gems:       []string{`gem "pg", "~> 1.5"`},
		springURL:  "jdbc:postgresql://%s:5432/%s",
		springDrv:  "org.postgresql.Driver",
	},
	"mysql": {
		nodeDeps:   map[string]string{"mysql2": "^3.9.8"},
		goRequires: []string{"gorm.io/gorm v1.25.10", "gorm.io/driver/mysql v1.5.7"},
		mavenDeps:  [][2]string{{"com.mysql", "mysql-connector-j"}},
		nugetPkgs:  [][2]string{{"Pomelo.EntityFrameworkCore.MySql", "8.0.2"}},
		cargoLines: []string{`sqlx = { version = "0.7", features = ["runtime-tokio", "mysql"] }`},
		gems:       []string{`gem "mysql2", "~> 0.5"`},
		springURL:  "jdbc:mysql://%s:3306/%s",
		springDrv:  "com.mysql.cj.jdbc.Driver",
	},
	"mongo": {
		nodeDeps:   map[string]string{"mongoose": "^8.4.1"},
		goRequires: []string{"go.mongodb.org/mongo-driver v1.15.0"},
		mavenDeps:  [][2]string{{"org.springframework.boot", "spring-boot-starter-data-mongodb"}},
		nugetPkgs:  [][2]string{{"MongoDB.Driver", "2.27.0"}},
		cargoLines: []string{`mongodb = "2.8"`},
		gems:       []string{`gem "mongoid", "~> 9.0"`},
	},
}

// InjectDriverManifest adds the client library for dbID to the framework's dependency manifest.
// It is idempotent and safe to call for any framework; missing manifests are skipped.
func InjectDriverManifest(cfg *config.ProjectConfig, dbID, host string) error {
	if dbID == "mariadb" {
		dbID = "mysql"
	}
	spec, ok := driverSpecs[dbID]
	if !ok {
		return nil
	}
	base := cfg.TargetPath

	switch cfg.Framework.ID {
	case "express", "nestjs", "next", "nuxt", "sveltekit", "astro", "vite-react", "vite-vue":
		if cfg.Framework.ID == "next" && dbID == "postgres" {
			return nil // Prisma handles Postgres for Next.js
		}
		addNodeDeps(filepath.Join(base, "package.json"), spec.nodeDeps)

	case "fiber", "gin":
		addGoRequires(filepath.Join(base, "go.mod"), spec.goRequires)

	case "springboot":
		addMavenDeps(filepath.Join(base, "pom.xml"), spec.mavenDeps)
		if spec.springURL != "" {
			props := filepath.Join(base, "src", "main", "resources", "application.properties")
			if b, err := os.ReadFile(props); err == nil && !strings.Contains(string(b), "spring.datasource.url") {
				add := fmt.Sprintf("\nspring.datasource.url=${DATABASE_URL:"+spec.springURL+"}\n"+
					"spring.datasource.username=${DB_USER:ace_user}\n"+
					"spring.datasource.password=${DB_PASSWORD:secret}\n"+
					"spring.datasource.driver-class-name=%s\n", host, cfg.DBName(), spec.springDrv)
				_ = os.WriteFile(props, append(b, []byte(add)...), 0644)
			}
		}

	case "aspnet":
		matches, _ := filepath.Glob(filepath.Join(base, "*.csproj"))
		for _, cs := range matches {
			b, err := os.ReadFile(cs)
			if err != nil {
				continue
			}
			s := string(b)
			for _, p := range spec.nugetPkgs {
				if !strings.Contains(s, p[0]) {
					s = strings.Replace(s, "</ItemGroup>",
						fmt.Sprintf("  <PackageReference Include=\"%s\" Version=\"%s\" />\n  </ItemGroup>", p[0], p[1]), 1)
				}
			}
			_ = os.WriteFile(cs, []byte(s), 0644)
		}

	case "axum":
		cargo := filepath.Join(base, "Cargo.toml")
		if b, err := os.ReadFile(cargo); err == nil {
			s := string(b)
			for _, l := range spec.cargoLines {
				key := strings.SplitN(l, " ", 2)[0]
				if !strings.Contains(s, key+" =") {
					s += "\n" + l + "\n"
				}
			}
			_ = os.WriteFile(cargo, []byte(s), 0644)
		}

	case "django", "fastapi":
		reqPath := filepath.Join(base, "requirements.txt")
		if b, err := os.ReadFile(reqPath); err == nil {
			s := string(b)
			switch dbID {
			case "mongo":
				if !strings.Contains(s, "pymongo") && !strings.Contains(s, "motor") {
					if cfg.Framework.ID == "fastapi" {
						s += "motor>=3.4.0\n"
					} else {
						s += "pymongo>=4.6.0\n"
					}
					_ = os.WriteFile(reqPath, []byte(s), 0644)
				}
			case "postgres":
				if !strings.Contains(s, "psycopg") && !strings.Contains(s, "asyncpg") {
					s += "psycopg[binary]>=3.1.0\n"
					_ = os.WriteFile(reqPath, []byte(s), 0644)
				}
			case "mysql":
				if !strings.Contains(s, "pymysql") && !strings.Contains(s, "aiomysql") {
					if cfg.Framework.ID == "fastapi" {
						s += "aiomysql>=0.2.0\n"
					} else {
						s += "pymysql>=1.1.0\n"
					}
					_ = os.WriteFile(reqPath, []byte(s), 0644)
				}
			}
		}

	case "rails":
		gemfile := filepath.Join(base, "Gemfile")
		if b, err := os.ReadFile(gemfile); err == nil {
			s := string(b)
			for _, g := range spec.gems {
				name := strings.Split(g, `"`)[1]
				if !strings.Contains(s, `"`+name+`"`) {
					s += "\n" + g + "\n"
				}
			}
			_ = os.WriteFile(gemfile, []byte(s), 0644)
		}
	}
	return nil
}

func addNodeDeps(pkgPath string, deps map[string]string) {
	b, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}
	s := string(b)
	if !strings.Contains(s, `"dependencies": {`) {
		if strings.Contains(s, `"devDependencies": {`) {
			s = strings.Replace(s, `"devDependencies": {`, "\"dependencies\": {\n  },\n  \"devDependencies\": {", 1)
		} else {
			last := strings.LastIndex(s, "}")
			if last != -1 {
				s = s[:last] + ",\n  \"dependencies\": {\n  }\n}"
			}
		}
	}
	for name, ver := range deps {
		if strings.Contains(s, `"`+name+`"`) {
			continue
		}
		if strings.Contains(s, `"dependencies": {}`) {
			s = strings.Replace(s, `"dependencies": {}`, fmt.Sprintf("\"dependencies\": {\n    \"%s\": \"%s\"\n  }", name, ver), 1)
		} else if strings.Contains(s, "\"dependencies\": {\n  }") {
			s = strings.Replace(s, "\"dependencies\": {\n  }", fmt.Sprintf("\"dependencies\": {\n    \"%s\": \"%s\"\n  }", name, ver), 1)
		} else {
			s = strings.Replace(s, `"dependencies": {`,
				fmt.Sprintf("\"dependencies\": {\n    \"%s\": \"%s\",", name, ver), 1)
		}
	}
	_ = os.WriteFile(pkgPath, []byte(s), 0644)
}

func addGoRequires(modPath string, reqs []string) {
	b, err := os.ReadFile(modPath)
	if err != nil {
		return
	}
	s := string(b)
	for _, r := range reqs {
		mod := strings.Fields(r)[0]
		if strings.Contains(s, mod+" ") {
			continue
		}
		s = strings.Replace(s, "require (", "require (\n\t"+r, 1)
	}
	_ = os.WriteFile(modPath, []byte(s), 0644)
}

func addMavenDeps(pomPath string, deps [][2]string) {
	b, err := os.ReadFile(pomPath)
	if err != nil {
		return
	}
	s := string(b)
	for _, d := range deps {
		if strings.Contains(s, "<artifactId>"+d[1]+"</artifactId>") {
			continue
		}
		s = strings.Replace(s, "<dependencies>", fmt.Sprintf(`<dependencies>
		<dependency>
			<groupId>%s</groupId>
			<artifactId>%s</artifactId>
		</dependency>`, d[0], d[1]), 1)
	}
	_ = os.WriteFile(pomPath, []byte(s), 0644)
}
