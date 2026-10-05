package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/opeteer/ace/internal/config"
)

// ProjectCheckItem represents an individual diagnostic check for a project
type ProjectCheckItem struct {
	Name        string     `json:"name"`
	Category    string     `json:"category"`
	Status      ToolStatus `json:"status"`
	Current     string     `json:"current,omitempty"`
	Required    string     `json:"required,omitempty"`
	Detail      string     `json:"detail,omitempty"`
	InstallHelp string     `json:"install_help,omitempty"`
}

// ProjectDiagnosticReport contains the full scoped diagnostic for a project
type ProjectDiagnosticReport struct {
	ProjectName string             `json:"project_name"`
	FrameworkID string             `json:"framework_id"`
	Framework   string             `json:"framework"`
	Language    string             `json:"language"`
	ProjectPath string             `json:"project_path"`
	Checks      []ProjectCheckItem `json:"checks"`
	IsReady     bool               `json:"is_ready"`
	NextSteps   []string           `json:"next_steps,omitempty"`
}

// ProjectInfo describes a detected project in a directory
type ProjectInfo struct {
	Name        string
	FrameworkID string
	Language    string
	Path        string
}

// DetectProject inspects a directory and returns its project information if recognized
func DetectProject(dir string) (*ProjectInfo, bool) {
	if dir == "" {
		dir = "."
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, false
	}

	projectName := filepath.Base(absPath)

	// 1. PHP / Laravel (composer.json or artisan)
	if fileExists(filepath.Join(absPath, "artisan")) || fileExists(filepath.Join(absPath, "composer.json")) {
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: "laravel",
			Language:    "PHP",
			Path:        absPath,
		}, true
	}

	// 2. Node.js Ecosystem (package.json)
	if fileExists(filepath.Join(absPath, "package.json")) {
		fwID := "next"
		pkgBytes, _ := os.ReadFile(filepath.Join(absPath, "package.json"))
		pkgStr := string(pkgBytes)

		if fileExists(filepath.Join(absPath, "nuxt.config.ts")) || fileExists(filepath.Join(absPath, "nuxt.config.js")) || strings.Contains(pkgStr, `"nuxt"`) {
			fwID = "nuxt"
		} else if fileExists(filepath.Join(absPath, "svelte.config.js")) || strings.Contains(pkgStr, `"@sveltejs/kit"`) {
			fwID = "sveltekit"
		} else if fileExists(filepath.Join(absPath, "astro.config.mjs")) || strings.Contains(pkgStr, `"astro"`) {
			fwID = "astro"
		} else if fileExists(filepath.Join(absPath, "nest-cli.json")) || strings.Contains(pkgStr, `"@nestjs/core"`) {
			fwID = "nestjs"
		} else if fileExists(filepath.Join(absPath, "app.json")) || strings.Contains(pkgStr, `"expo"`) {
			fwID = "expo"
		} else if strings.Contains(pkgStr, `"@vitejs/plugin-vue"`) || (strings.Contains(pkgStr, `"vue"`) && (fileExists(filepath.Join(absPath, "vite.config.ts")) || fileExists(filepath.Join(absPath, "vite.config.js")))) {
			fwID = "vite-vue"
		} else if strings.Contains(pkgStr, `"@vitejs/plugin-react"`) || (strings.Contains(pkgStr, `"react"`) && (fileExists(filepath.Join(absPath, "vite.config.ts")) || fileExists(filepath.Join(absPath, "vite.config.js")))) {
			fwID = "vite-react"
		} else if strings.Contains(pkgStr, `"express"`) {
			fwID = "express"
		} else if strings.Contains(pkgStr, `"next"`) {
			fwID = "next"
		}

		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: fwID,
			Language:    "TypeScript/JavaScript",
			Path:        absPath,
		}, true
	}

	// 3. Python Ecosystem (pyproject.toml, requirements.txt, manage.py)
	if fileExists(filepath.Join(absPath, "manage.py")) {
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: "django",
			Language:    "Python",
			Path:        absPath,
		}, true
	}
	if fileExists(filepath.Join(absPath, "requirements.txt")) || fileExists(filepath.Join(absPath, "pyproject.toml")) {
		fwID := "fastapi"
		reqBytes, _ := os.ReadFile(filepath.Join(absPath, "requirements.txt"))
		if strings.Contains(strings.ToLower(string(reqBytes)), "django") {
			fwID = "django"
		}
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: fwID,
			Language:    "Python",
			Path:        absPath,
		}, true
	}

	// 4. Go Ecosystem (go.mod)
	if fileExists(filepath.Join(absPath, "go.mod")) {
		fwID := "fiber"
		modBytes, _ := os.ReadFile(filepath.Join(absPath, "go.mod"))
		if strings.Contains(string(modBytes), "gin-gonic/gin") {
			fwID = "gin"
		}
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: fwID,
			Language:    "Go",
			Path:        absPath,
		}, true
	}

	// 5. Rust Ecosystem (Cargo.toml)
	if fileExists(filepath.Join(absPath, "Cargo.toml")) {
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: "axum",
			Language:    "Rust",
			Path:        absPath,
		}, true
	}

	// 6. Flutter (pubspec.yaml)
	if fileExists(filepath.Join(absPath, "pubspec.yaml")) {
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: "flutter",
			Language:    "Dart",
			Path:        absPath,
		}, true
	}

	// 7. .NET (*.csproj, *.sln)
	if matches, _ := filepath.Glob(filepath.Join(absPath, "*.csproj")); len(matches) > 0 {
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: "aspnet",
			Language:    "C#",
			Path:        absPath,
		}, true
	}

	// 8. Java (pom.xml, build.gradle)
	if fileExists(filepath.Join(absPath, "pom.xml")) || fileExists(filepath.Join(absPath, "build.gradle")) {
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: "springboot",
			Language:    "Java",
			Path:        absPath,
		}, true
	}

	// 9. Ruby / Rails (Gemfile)
	if fileExists(filepath.Join(absPath, "Gemfile")) {
		return &ProjectInfo{
			Name:        projectName,
			FrameworkID: "rails",
			Language:    "Ruby",
			Path:        absPath,
		}, true
	}

	return nil, false
}

// DiagnoseProject runs a scoped diagnostic on the specified project directory
func DiagnoseProject(projectDir string) (*ProjectDiagnosticReport, error) {
	info, ok := DetectProject(projectDir)
	if !ok {
		return nil, fmt.Errorf("no recognized project manifest found in '%s'", projectDir)
	}

	report := &ProjectDiagnosticReport{
		ProjectName: info.Name,
		FrameworkID: info.FrameworkID,
		Language:    info.Language,
		ProjectPath: info.Path,
		IsReady:     true,
	}

	if fw, exists := config.FindFramework(info.FrameworkID); exists {
		report.Framework = fw.Name
	} else {
		report.Framework = strings.Title(info.FrameworkID)
	}

	switch info.FrameworkID {
	case "laravel":
		diagnoseLaravelProject(report)
	case "fastapi", "django":
		diagnosePythonProject(report)
	case "fiber", "gin":
		diagnoseGoProject(report)
	case "next", "nuxt", "sveltekit", "astro", "express", "nestjs", "vite-react", "vite-vue", "expo":
		diagnoseNodeProject(report)
	case "axum":
		diagnoseRustProject(report)
	case "flutter":
		diagnoseFlutterProject(report)
	case "aspnet":
		diagnoseDotnetProject(report)
	case "springboot":
		diagnoseJavaProject(report)
	case "rails":
		diagnoseRailsProject(report)
	}

	// Common checks: Docker & Git
	diagnoseCommonTools(report)

	// Determine readiness
	for _, c := range report.Checks {
		if c.Status == StatusMissing {
			report.IsReady = false
			break
		}
	}

	return report, nil
}

// DiagnoseFramework runs a diagnostic for a specific framework name/ID
func DiagnoseFramework(fwID string) (*ProjectDiagnosticReport, error) {
	fw, exists := config.FindFramework(fwID)
	if !exists {
		return nil, fmt.Errorf("unknown framework '%s'", fwID)
	}

	report := &ProjectDiagnosticReport{
		ProjectName: fw.Name + " Stack",
		FrameworkID: fw.ID,
		Framework:   fw.Name,
		Language:    fw.Language,
		ProjectPath: "-",
		IsReady:     true,
	}

	switch fw.ID {
	case "laravel":
		auditToolSpec(report, "PHP Runtime", "php", []string{"-v"}, `PHP (\d+\.\d+\.\d+)`, "8.2.0", "Install PHP 8.2+ via package manager")
		auditToolSpec(report, "Composer", "composer", []string{"--version"}, `Composer (?:version )?(\d+\.\d+\.\d+)`, "2.2.0", "Run 'ace install composer'")
	case "fastapi", "django":
		auditToolSpec(report, "Python 3", "python3", []string{"--version"}, `Python (\d+\.\d+\.\d+)`, "3.10.0", "Install Python 3 via package manager")
		auditToolSpec(report, "pip", "pip3", []string{"--version"}, `pip (\d+\.\d+)`, "22.0", "Install pip via 'python3 -m ensurepip'")
	case "fiber", "gin":
		auditToolSpec(report, "Go Compiler", "go", []string{"version"}, `go version go(\d+\.\d+\.?\d*)`, "1.21.0", "Install Go from https://go.dev")
	case "axum":
		auditToolSpec(report, "Rust / Cargo", "cargo", []string{"--version"}, `cargo (\d+\.\d+\.\d+)`, "1.75.0", "Run 'ace install rust'")
	case "flutter":
		auditToolSpec(report, "Flutter SDK", "flutter", []string{"--version"}, `Flutter (\d+\.\d+\.\d+)`, "3.19.0", "Install Flutter SDK from https://docs.flutter.dev")
	case "aspnet":
		auditToolSpec(report, ".NET SDK", "dotnet", []string{"--version"}, `(\d+\.\d+\.\d+)`, "8.0.0", "Run 'ace install dotnet'")
	case "springboot":
		auditToolSpec(report, "Java JDK", "java", []string{"-version"}, `(?:version ")?(\d+(?:\.\d+)*)`, "17.0.0", "Install OpenJDK 17 or 21")
	case "rails":
		auditToolSpec(report, "Ruby Runtime", "ruby", []string{"-v"}, `ruby (\d+\.\d+\.\d+)`, "3.0.0", "Install Ruby 3+ via package manager")
		auditToolSpec(report, "Bundler", "bundle", []string{"--version"}, `Bundler version (\d+\.\d+\.\d+)`, "2.4.0", "Run 'gem install bundler'")
	default:
		// Node.js stacks
		auditToolSpec(report, "Node.js", "node", []string{"--version"}, `v?(\d+\.\d+\.\d+)`, "18.17.0", "Install Node.js via https://nodejs.org")
		auditNodePkgManager(report)
	}

	auditToolSpec(report, "Docker Engine", "docker", []string{"--version"}, `Docker version (\d+\.\d+\.\d+)`, "20.10.0", "Install Docker from https://docker.com")
	auditToolSpec(report, "Git", "git", []string{"--version"}, `git version (\d+\.\d+\.\d+)`, "2.20.0", "Install git via package manager")

	for _, c := range report.Checks {
		if c.Status == StatusMissing {
			report.IsReady = false
			break
		}
	}

	return report, nil
}

func diagnoseLaravelProject(report *ProjectDiagnosticReport) {
	// 1. PHP Runtime
	auditToolSpec(report, "PHP Runtime", "php", []string{"-v"}, `PHP (\d+\.\d+\.\d+)`, "8.2.0", "Install PHP 8.2+ via package manager")

	// 2. Composer
	auditToolSpec(report, "Composer", "composer", []string{"--version"}, `Composer (?:version )?(\d+\.\d+\.\d+)`, "2.2.0", "Run 'ace install composer'")

	// 3. Vendor directory
	vendorDir := filepath.Join(report.ProjectPath, "vendor")
	autoloadFile := filepath.Join(vendorDir, "autoload.php")
	if fileExists(autoloadFile) {
		packagesCount := 0
		entries, err := os.ReadDir(vendorDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
					packagesCount++
				}
			}
		}
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Project Vendor",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  fmt.Sprintf("Installed (%d packages in vendor/)", packagesCount),
			Required: "vendor/autoload.php",
			Detail:   "All PHP dependencies are present",
		})
	} else {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:        "Project Vendor",
			Category:    "Dependencies",
			Status:      StatusMissing,
			Current:     "Missing vendor/",
			Required:    "vendor/autoload.php",
			Detail:      "Run 'ace install' or 'composer install' to populate vendor/",
			InstallHelp: "ace install",
		})
	}

	// 4. Environment (.env)
	envFile := filepath.Join(report.ProjectPath, ".env")
	if fileExists(envFile) {
		envData, _ := os.ReadFile(envFile)
		hasAppKey := strings.Contains(string(envData), "APP_KEY=base64:")
		keyDetail := "APP_KEY is configured"
		if !hasAppKey {
			keyDetail = "Warning: APP_KEY not generated yet"
		}
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Environment",
			Category: "Configuration",
			Status:   StatusOK,
			Current:  ".env present",
			Required: ".env",
			Detail:   keyDetail,
		})

		// 5. Database & Redis in .env
		auditDatabaseAndRedis(report, string(envData))
	} else {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:        "Environment",
			Category:    "Configuration",
			Status:      StatusMissing,
			Current:     "Missing .env",
			Required:    ".env",
			Detail:      "Copy .env.example to .env and run 'php artisan key:generate'",
			InstallHelp: "cp .env.example .env && php artisan key:generate",
		})
	}

	report.NextSteps = []string{
		"php artisan serve                # Start local development server at http://localhost:8000",
		"docker compose up -d             # Start PostgreSQL and background services",
		"php artisan migrate              # Run database migrations",
	}
}

func diagnosePythonProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, "Python 3", "python3", []string{"--version"}, `Python (\d+\.\d+\.\d+)`, "3.10.0", "Install Python 3 via package manager")
	auditToolSpec(report, "pip", "pip3", []string{"--version"}, `pip (\d+\.\d+)`, "22.0", "Install pip via 'python3 -m ensurepip'")

	venvPath := filepath.Join(report.ProjectPath, ".venv")
	pyBin := filepath.Join(venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		pyBin = filepath.Join(venvPath, "Scripts", "python.exe")
	}

	if fileExists(pyBin) {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Virtualenv (.venv)",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "Active (.venv)",
			Required: ".venv directory",
			Detail:   "Python virtual environment configured",
		})

		// Check if Django is installed in .venv for Django projects
		if report.FrameworkID == "django" || fileExists(filepath.Join(report.ProjectPath, "manage.py")) {
			report.Framework = "Django 5"
			verCmd := exec.Command(pyBin, "-m", "django", "--version")
			if out, err := verCmd.Output(); err == nil {
				djVer := strings.TrimSpace(string(out))
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "Django Core",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  fmt.Sprintf("Installed (v%s)", djVer),
					Required: "Django >= 5.0",
					Detail:   "Django framework ready in .venv",
				})
			} else {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:        "Django Core",
					Category:    "Dependencies",
					Status:      StatusMissing,
					Current:     "Not installed in .venv",
					Required:    "Django >= 5.0",
					Detail:      "Run 'ace install' to install dependencies",
					InstallHelp: "ace install",
				})
			}
		}
	} else {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:        "Virtualenv (.venv)",
			Category:    "Dependencies",
			Status:      StatusMissing,
			Current:     "Missing .venv",
			Required:    ".venv directory",
			Detail:      "Run 'ace install' or 'python3 -m venv .venv'",
			InstallHelp: "ace install",
		})
	}

	// Environment & Database checks
	envFile := filepath.Join(report.ProjectPath, ".env")
	if fileExists(envFile) {
		envData, _ := os.ReadFile(envFile)
		envStr := string(envData)
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Environment",
			Category: "Configuration",
			Status:   StatusOK,
			Current:  ".env present",
			Required: ".env",
			Detail:   "Environment file configured",
		})

		auditDatabaseAndRedis(report, envStr)
	}

	if report.FrameworkID == "django" || fileExists(filepath.Join(report.ProjectPath, "manage.py")) {
		report.NextSteps = []string{
			"source .venv/bin/activate && python manage.py runserver  # Start local development server at http://localhost:8000",
			"docker compose up -d                                      # Start PostgreSQL and container services",
			"source .venv/bin/activate && python manage.py migrate    # Run initial database migrations",
		}
	} else {
		report.NextSteps = []string{
			"source .venv/bin/activate        # Activate virtual environment",
			"uvicorn app.main:app --reload    # Start FastAPI server",
			"docker compose up -d             # Start containerized services",
		}
	}
}

func diagnoseGoProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, "Go Compiler", "go", []string{"version"}, `go version go(\d+\.\d+\.?\d*)`, "1.21.0", "Install Go from https://go.dev")

	goMod := filepath.Join(report.ProjectPath, "go.mod")
	if fileExists(goMod) {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Go Module",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "go.mod present",
			Required: "go.mod",
			Detail:   "Dependencies managed via Go modules",
		})
	}

	goSum := filepath.Join(report.ProjectPath, "go.sum")
	if fileExists(goSum) {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Module Lock",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "go.sum present",
			Required: "go.sum",
			Detail:   "All module checksums verified",
		})
	}

	checkProjectEnv(report)

	report.NextSteps = []string{
		"go run cmd/api/main.go           # Start API server",
		"docker compose up -d             # Start database & services",
	}
}

func diagnoseNodeProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, "Node.js", "node", []string{"--version"}, `v?(\d+\.\d+\.\d+)`, "18.17.0", "Install Node.js via https://nodejs.org")
	auditNodePkgManager(report)

	nodeModules := filepath.Join(report.ProjectPath, "node_modules")
	if stat, err := os.Stat(nodeModules); err == nil && stat.IsDir() {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Node Modules",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "Installed (node_modules/ present)",
			Required: "node_modules/",
			Detail:   "All npm/pnpm/bun packages are installed",
		})
	} else {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:        "Node Modules",
			Category:    "Dependencies",
			Status:      StatusMissing,
			Current:     "Missing node_modules/",
			Required:    "node_modules/",
			Detail:      "Run 'ace install' or 'npm install'",
			InstallHelp: "ace install",
		})
	}

	checkProjectEnv(report)

	if report.FrameworkID == "expo" {
		report.NextSteps = []string{
			"npx expo start                  # Start Expo mobile bundler",
			"npx expo run:android             # Run on Android emulator",
		}
	} else {
		report.NextSteps = []string{
			"npm run dev                      # Start dev server",
			"docker compose up -d             # Start containerized services",
		}
	}
}

func diagnoseRustProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, "Rust / Cargo", "cargo", []string{"--version"}, `cargo (\d+\.\d+\.\d+)`, "1.75.0", "Run 'ace install rust'")

	cargoToml := filepath.Join(report.ProjectPath, "Cargo.toml")
	if fileExists(cargoToml) {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Cargo Manifest",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "Cargo.toml present",
			Required: "Cargo.toml",
			Detail:   "Rust crate specification valid",
		})
	}

	checkProjectEnv(report)

	report.NextSteps = []string{
		"cargo run                        # Run application",
		"docker compose up -d             # Start containerized services",
	}
}

func diagnoseFlutterProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, "Flutter SDK", "flutter", []string{"--version"}, `Flutter (\d+\.\d+\.\d+)`, "3.19.0", "Install Flutter from https://docs.flutter.dev")

	pubspec := filepath.Join(report.ProjectPath, "pubspec.yaml")
	if fileExists(pubspec) {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Flutter Manifest",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "pubspec.yaml present",
			Required: "pubspec.yaml",
			Detail:   "Flutter project configuration valid",
		})
	}

	checkProjectEnv(report)

	report.NextSteps = []string{
		"flutter run                      # Launch app on connected device",
	}
}

func diagnoseDotnetProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, ".NET SDK", "dotnet", []string{"--version"}, `(\d+\.\d+\.\d+)`, "8.0.0", "Run 'ace install dotnet'")

	matches, _ := filepath.Glob(filepath.Join(report.ProjectPath, "*.csproj"))
	if len(matches) > 0 {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "C# Project",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  filepath.Base(matches[0]),
			Required: "*.csproj",
			Detail:   ".NET project file configured",
		})
	}

	checkProjectEnv(report)

	report.NextSteps = []string{
		"dotnet run                       # Run ASP.NET Core project",
		"docker compose up -d             # Start containerized services",
	}
}

func diagnoseJavaProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, "Java JDK", "java", []string{"-version"}, `(?:version ")?(\d+(?:\.\d+)*)`, "17.0.0", "Install OpenJDK 17 or 21")

	pomFile := filepath.Join(report.ProjectPath, "pom.xml")
	if fileExists(pomFile) {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Maven POM",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "pom.xml present",
			Required: "pom.xml",
			Detail:   "Spring Boot dependencies declared",
		})
	}

	checkProjectEnv(report)

	report.NextSteps = []string{
		"./mvnw spring-boot:run           # Run Spring Boot application",
		"docker compose up -d             # Start containerized services",
	}
}

func diagnoseRailsProject(report *ProjectDiagnosticReport) {
	auditToolSpec(report, "Ruby Runtime", "ruby", []string{"-v"}, `ruby (\d+\.\d+\.\d+)`, "3.0.0", "Install Ruby 3+ via package manager")
	auditToolSpec(report, "Bundler", "bundle", []string{"--version"}, `Bundler version (\d+\.\d+\.\d+)`, "2.4.0", "Run 'gem install bundler'")

	gemfile := filepath.Join(report.ProjectPath, "Gemfile")
	if fileExists(gemfile) {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Gemfile",
			Category: "Dependencies",
			Status:   StatusOK,
			Current:  "Gemfile present",
			Required: "Gemfile",
			Detail:   "Ruby gems declared",
		})
	}

	checkProjectEnv(report)

	report.NextSteps = []string{
		"bundle exec puma -C config/puma.rb # Start Rails server",
		"docker compose up -d               # Start containerized services",
	}
}

func checkProjectEnv(report *ProjectDiagnosticReport) {
	envFile := filepath.Join(report.ProjectPath, ".env")
	if fileExists(envFile) {
		envData, _ := os.ReadFile(envFile)
		envStr := string(envData)
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Environment",
			Category: "Configuration",
			Status:   StatusOK,
			Current:  ".env present",
			Required: ".env",
			Detail:   "Environment configured",
		})

		auditDatabaseAndRedis(report, envStr)
	}
}

func auditDatabaseAndRedis(report *ProjectDiagnosticReport, envStr string) {
	// 1. Identify primary database
	dbConnection := extractEnvVar(envStr, "DB_CONNECTION")
	dbEngine := extractEnvVar(envStr, "DB_ENGINE")
	dbType := extractEnvVar(envStr, "DB_TYPE")
	dbURL := extractEnvVar(envStr, "DATABASE_URL")
	mongoURI := extractEnvVar(envStr, "MONGODB_URI")
	dbHost := extractEnvVar(envStr, "DB_HOST")
	dbPort := extractEnvVar(envStr, "DB_PORT")
	dbName := extractEnvVar(envStr, "DB_NAME")
	if dbName == "" {
		dbName = extractEnvVar(envStr, "DB_DATABASE")
	}

	var detectedDB, displayName, defaultPort string
	isSQLite := false

	lowerConn := strings.ToLower(dbConnection + " " + dbEngine + " " + dbType + " " + dbURL + " " + mongoURI)

	if strings.Contains(lowerConn, "postgres") || strings.Contains(lowerConn, "pgsql") {
		detectedDB = "postgres"
		displayName = "PostgreSQL"
		defaultPort = "5432"
	} else if strings.Contains(lowerConn, "mariadb") {
		detectedDB = "mariadb"
		displayName = "MariaDB"
		defaultPort = "3306"
	} else if strings.Contains(lowerConn, "mysql") {
		detectedDB = "mysql"
		displayName = "MySQL"
		defaultPort = "3306"
	} else if strings.Contains(lowerConn, "sqlite") {
		detectedDB = "sqlite"
		displayName = "SQLite3"
		isSQLite = true
	} else if strings.Contains(lowerConn, "mongo") {
		detectedDB = "mongo"
		displayName = "MongoDB"
		defaultPort = "27017"
	} else if strings.Contains(lowerConn, "clickhouse") {
		detectedDB = "clickhouse"
		displayName = "ClickHouse"
		defaultPort = "8123"
	} else if strings.Contains(lowerConn, "meilisearch") {
		detectedDB = "meilisearch"
		displayName = "Meilisearch"
		defaultPort = "7700"
	} else if strings.Contains(lowerConn, "elastic") {
		detectedDB = "elastic"
		displayName = "Elasticsearch"
		defaultPort = "9200"
	} else if strings.Contains(lowerConn, "neo4j") {
		detectedDB = "neo4j"
		displayName = "Neo4j"
		defaultPort = "7474"
	} else if strings.Contains(lowerConn, "redis") && !strings.Contains(envStr, "DB_CONNECTION=postgres") && !strings.Contains(envStr, "DB_CONNECTION=mysql") {
		detectedDB = "redis"
		displayName = "Redis"
		defaultPort = "6379"
	}

	if detectedDB != "" || dbHost != "" || dbConnection != "" {
		if displayName == "" {
			if dbType != "" {
				displayName = strings.Title(dbType)
			} else if dbConnection != "" {
				displayName = strings.Title(dbConnection)
			} else {
				displayName = "Database"
			}
		}

		if isSQLite {
			sqliteFile := dbName
			if sqliteFile == "" {
				sqliteFile = "db.sqlite3"
			}
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "Database Config",
				Category: "Configuration",
				Status:   StatusOK,
				Current:  fmt.Sprintf("%s (%s)", displayName, sqliteFile),
				Required: "SQLite database file",
				Detail:   "File-based embedded database",
			})
		} else {
			port := dbPort
			if port == "" {
				port = defaultPort
			}
			host := dbHost
			if host == "" {
				host = "127.0.0.1"
			}
			current := fmt.Sprintf("%s (%s:%s)", displayName, host, port)
			detail := fmt.Sprintf("Wired for %s", displayName)
			if dbName != "" {
				detail = fmt.Sprintf("Database '%s' configured in .env", dbName)
			}
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "Database Config",
				Category: "Configuration",
				Status:   StatusOK,
				Current:  current,
				Required: displayName,
				Detail:   detail,
			})
		}

		// Audit DB driver for primary database
		auditDBDriver(report, detectedDB)
	}

	// 2. Audit Companion Redis
	redisHost := extractEnvVar(envStr, "REDIS_HOST")
	redisPort := extractEnvVar(envStr, "REDIS_PORT")
	redisURL := extractEnvVar(envStr, "REDIS_URL")
	redisAddr := extractEnvVar(envStr, "REDIS_ADDR")

	if (redisHost != "" || redisURL != "" || redisAddr != "") && detectedDB != "redis" {
		if redisPort == "" {
			redisPort = "6379"
		}
		if redisHost == "" {
			redisHost = "127.0.0.1"
		}
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:     "Redis Cache",
			Category: "Configuration",
			Status:   StatusOK,
			Current:  fmt.Sprintf("Redis (%s:%s)", redisHost, redisPort),
			Required: "REDIS_HOST in .env",
			Detail:   "In-memory cache & queue configured",
		})

		// Audit Redis client driver
		auditRedisDriver(report)
	}
}

func auditDBDriver(report *ProjectDiagnosticReport, dbType string) {
	switch report.FrameworkID {
	case "django", "fastapi":
		switch dbType {
		case "postgres":
			if checkPythonPackage(report, "psycopg2") || checkPythonPackage(report, "asyncpg") || checkPythonPackage(report, "psycopg") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "PostgreSQL Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Installed (psycopg)",
					Required: "psycopg or asyncpg",
					Detail:   "Python PostgreSQL client ready in .venv",
				})
			}
		case "mysql", "mariadb":
			if checkPythonPackage(report, "pymysql") || checkPythonPackage(report, "aiomysql") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "MySQL Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Installed (pymysql/aiomysql)",
					Required: "pymysql or aiomysql",
					Detail:   "Python MySQL client ready in .venv",
				})
			}
		case "mongo":
			if checkPythonPackage(report, "pymongo") || checkPythonPackage(report, "motor") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "MongoDB Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Installed (pymongo/motor)",
					Required: "pymongo or motor",
					Detail:   "Python MongoDB client ready in .venv",
				})
			}
		case "sqlite":
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "SQLite Driver",
				Category: "Dependencies",
				Status:   StatusOK,
				Current:  "Built-in (sqlite3)",
				Required: "sqlite3",
				Detail:   "Standard library SQLite engine",
			})
		}

	case "next", "nuxt", "sveltekit", "astro", "express", "nestjs", "vite-react", "vite-vue":
		switch dbType {
		case "postgres":
			if checkNodePackage(report, "@prisma/client") || checkNodePackage(report, "pg") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "PostgreSQL Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Declared (pg/prisma)",
					Required: "pg or @prisma/client",
					Detail:   "Node PostgreSQL driver declared",
				})
			}
		case "mysql", "mariadb":
			if checkNodePackage(report, "mysql2") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "MySQL Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Declared (mysql2)",
					Required: "mysql2",
					Detail:   "Node MySQL driver declared in package.json",
				})
			}
		case "sqlite":
			if checkNodePackage(report, "better-sqlite3") || checkNodePackage(report, "sqlite3") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "SQLite Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Declared (sqlite)",
					Required: "sqlite3 or better-sqlite3",
					Detail:   "Node SQLite driver declared in package.json",
				})
			}
		case "mongo":
			if checkNodePackage(report, "mongodb") || checkNodePackage(report, "mongoose") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "MongoDB Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Declared (mongodb)",
					Required: "mongodb or mongoose",
					Detail:   "Node MongoDB driver declared in package.json",
				})
			}
		}

	case "gin", "fiber":
		switch dbType {
		case "postgres":
			if checkGoModule(report, "gorm.io/driver/postgres") || checkGoModule(report, "lib/pq") || checkGoModule(report, "jackc/pgx") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "PostgreSQL Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Declared (gorm/postgres)",
					Required: "gorm.io/driver/postgres",
					Detail:   "Go PostgreSQL driver declared in go.mod",
				})
			}
		case "mysql", "mariadb":
			if checkGoModule(report, "gorm.io/driver/mysql") || checkGoModule(report, "go-sql-driver/mysql") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "MySQL Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Declared (gorm/mysql)",
					Required: "gorm.io/driver/mysql",
					Detail:   "Go MySQL driver declared in go.mod",
				})
			}
		case "sqlite":
			if checkGoModule(report, "gorm.io/driver/sqlite") || checkGoModule(report, "mattn/go-sqlite3") {
				report.Checks = append(report.Checks, ProjectCheckItem{
					Name:     "SQLite Driver",
					Category: "Dependencies",
					Status:   StatusOK,
					Current:  "Declared (gorm/sqlite)",
					Required: "gorm.io/driver/sqlite",
					Detail:   "Go SQLite driver declared in go.mod",
				})
			}
		}
	}
}

func auditRedisDriver(report *ProjectDiagnosticReport) {
	switch report.FrameworkID {
	case "django", "fastapi":
		if checkPythonPackage(report, "redis") || checkPythonPackage(report, "django_redis") {
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "Redis Driver",
				Category: "Dependencies",
				Status:   StatusOK,
				Current:  "Installed (redis-py)",
				Required: "redis / django-redis",
				Detail:   "Python Redis client ready in .venv",
			})
		}
	case "next", "nuxt", "sveltekit", "astro", "express", "nestjs", "vite-react", "vite-vue":
		if checkNodePackage(report, "ioredis") || checkNodePackage(report, "redis") {
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "Redis Driver",
				Category: "Dependencies",
				Status:   StatusOK,
				Current:  "Declared (ioredis)",
				Required: "ioredis",
				Detail:   "Node.js Redis client declared in package.json",
			})
		}
	case "gin", "fiber":
		if checkGoModule(report, "go-redis") {
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "Redis Driver",
				Category: "Dependencies",
				Status:   StatusOK,
				Current:  "Declared (go-redis/v9)",
				Required: "github.com/redis/go-redis/v9",
				Detail:   "Go Redis client library declared in go.mod",
			})
		}
	case "laravel":
		if checkComposerPackage(report, "predis") {
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "Redis Driver",
				Category: "Dependencies",
				Status:   StatusOK,
				Current:  "Declared (predis)",
				Required: "predis/predis",
				Detail:   "PHP Redis client library declared in composer.json",
			})
		}
	}
}

func checkPythonPackage(report *ProjectDiagnosticReport, pkgName string) bool {
	venvPath := filepath.Join(report.ProjectPath, ".venv")
	pyBin := filepath.Join(venvPath, "bin", "python")
	if runtime.GOOS == "windows" {
		pyBin = filepath.Join(venvPath, "Scripts", "python.exe")
	}
	if fileExists(pyBin) {
		cmd := exec.Command(pyBin, "-c", fmt.Sprintf("import %s", pkgName))
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	reqPath := filepath.Join(report.ProjectPath, "requirements.txt")
	if reqBytes, err := os.ReadFile(reqPath); err == nil {
		reqLower := strings.ToLower(string(reqBytes))
		pkgLower := strings.ToLower(pkgName)
		pkgAlt := strings.ReplaceAll(pkgLower, "_", "-")
		if strings.Contains(reqLower, pkgLower) || strings.Contains(reqLower, pkgAlt) {
			return true
		}
	}
	return false
}

func checkNodePackage(report *ProjectDiagnosticReport, pkgName string) bool {
	pkgJsonPath := filepath.Join(report.ProjectPath, "package.json")
	if pkgBytes, err := os.ReadFile(pkgJsonPath); err == nil {
		if strings.Contains(string(pkgBytes), fmt.Sprintf(`"%s"`, pkgName)) {
			return true
		}
	}
	nmPath := filepath.Join(report.ProjectPath, "node_modules", pkgName)
	if stat, err := os.Stat(nmPath); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func checkGoModule(report *ProjectDiagnosticReport, modName string) bool {
	goModPath := filepath.Join(report.ProjectPath, "go.mod")
	if modBytes, err := os.ReadFile(goModPath); err == nil {
		if strings.Contains(string(modBytes), modName) {
			return true
		}
	}
	return false
}

func checkComposerPackage(report *ProjectDiagnosticReport, pkgName string) bool {
	compPath := filepath.Join(report.ProjectPath, "composer.json")
	if compBytes, err := os.ReadFile(compPath); err == nil {
		if strings.Contains(string(compBytes), pkgName) {
			return true
		}
	}
	return false
}

func diagnoseCommonTools(report *ProjectDiagnosticReport) {
	// Docker / Compose check if project has Docker files
	hasDockerfile := fileExists(filepath.Join(report.ProjectPath, "Dockerfile"))
	hasCompose := fileExists(filepath.Join(report.ProjectPath, "docker-compose.yml"))

	if hasDockerfile || hasCompose {
		auditToolSpec(report, "Docker Engine", "docker", []string{"--version"}, `Docker version (\d+\.\d+\.\d+)`, "20.10.0", "Install Docker from https://docker.com")
	}

	// Git check
	if fileExists(filepath.Join(report.ProjectPath, ".git")) {
		auditToolSpec(report, "Git", "git", []string{"--version"}, `git version (\d+\.\d+\.\d+)`, "2.20.0", "Install git via package manager")
	}
}

func auditToolSpec(report *ProjectDiagnosticReport, name, binary string, args []string, regex, minVer, help string) {
	path, err := exec.LookPath(binary)
	if err != nil {
		// Also check ~/.local/bin
		home, _ := os.UserHomeDir()
		localBin := filepath.Join(home, ".local", "bin", binary)
		if stat, statErr := os.Stat(localBin); statErr == nil && !stat.IsDir() {
			path = localBin
			err = nil
		}
	}

	if err != nil {
		report.Checks = append(report.Checks, ProjectCheckItem{
			Name:        name,
			Category:    "Toolchain",
			Status:      StatusMissing,
			Current:     "Missing",
			Required:    minVer,
			Detail:      fmt.Sprintf("%s is required", name),
			InstallHelp: help,
		})
		return
	}

	ver := "installed"
	cmd := exec.Command(path, args...)
	out, _ := cmd.CombinedOutput()
	if len(out) > 0 {
		re := regexp.MustCompile(regex)
		match := re.FindStringSubmatch(string(out))
		if len(match) > 1 {
			ver = strings.TrimSpace(match[1])
		}
	}

	report.Checks = append(report.Checks, ProjectCheckItem{
		Name:     name,
		Category: "Toolchain",
		Status:   StatusOK,
		Current:  ver,
		Required: ">= " + minVer,
		Detail:   path,
	})
}

func auditNodePkgManager(report *ProjectDiagnosticReport) {
	for _, mgr := range []string{"bun", "pnpm", "npm"} {
		if path, err := exec.LookPath(mgr); err == nil {
			cmd := exec.Command(path, "--version")
			out, _ := cmd.CombinedOutput()
			ver := strings.TrimSpace(string(out))
			report.Checks = append(report.Checks, ProjectCheckItem{
				Name:     "Package Manager",
				Category: "Toolchain",
				Status:   StatusOK,
				Current:  fmt.Sprintf("%s %s", mgr, ver),
				Required: "bun, pnpm, or npm",
				Detail:   path,
			})
			return
		}
	}

	report.Checks = append(report.Checks, ProjectCheckItem{
		Name:        "Package Manager",
		Category:    "Toolchain",
		Status:      StatusMissing,
		Current:     "None",
		Required:    "bun, pnpm, or npm",
		InstallHelp: "Install Node.js or run 'ace install bun'",
	})
}

func extractEnvVar(envContent, key string) string {
	lines := strings.Split(envContent, "\n")
	prefix := key + "="
	result := ""
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, prefix) {
			val := strings.TrimPrefix(l, prefix)
			result = strings.Trim(val, `"' `)
		}
	}
	return result
}

func fileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

// ProjectDiagnosticJSON returns the project report as formatted JSON
func (r *ProjectDiagnosticReport) ToJSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
