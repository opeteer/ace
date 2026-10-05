package doctor

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
)

// ToolStatus represents the health check status of a tool
type ToolStatus string

const (
	StatusOK      ToolStatus = "OK"
	StatusMissing ToolStatus = "MISSING"
	StatusWarning ToolStatus = "WARN"
)

// ToolDiagnostic stores the diagnostic result for a single toolchain item
type ToolDiagnostic struct {
	Name           string     `json:"name"`
	Category       string     `json:"category"`
	Binary         string     `json:"binary"`
	Status         ToolStatus `json:"status"`
	Version        string     `json:"version,omitempty"`
	Path           string     `json:"path,omitempty"`
	MinVersion     string     `json:"min_version,omitempty"`
	InstallHelp    string     `json:"install_help,omitempty"`
	Provisionable  bool       `json:"provisionable"`
	UnlockedStacks []string   `json:"unlocked_stacks"`
}

// ToolSpec defines how to inspect a single tool
type ToolSpec struct {
	Name           string
	Category       string
	Binary         string
	VersionArgs    []string
	VersionRegex   string
	MinVersion     string
	InstallHelp    string
	UnlockedStacks []string
}

// Registry of tools to audit in `ace doctor`
var AuditedTools = []ToolSpec{
	// JavaScript / TypeScript Ecosystem
	{
		Name:           "Node.js",
		Category:       "JavaScript / TypeScript",
		Binary:         "node",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `v?(\d+\.\d+\.\d+)`,
		MinVersion:     "18.17.0",
		InstallHelp:    "Install via fnm, nvm, or https://nodejs.org",
		UnlockedStacks: []string{"next", "nuxt", "sveltekit", "astro", "express", "nestjs", "vite-react", "vite-vue", "expo"},
	},
	{
		Name:           "npm",
		Category:       "JavaScript / TypeScript",
		Binary:         "npm",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `(\d+\.\d+\.\d+)`,
		MinVersion:     "9.0.0",
		InstallHelp:    "Bundled with Node.js",
		UnlockedStacks: []string{"All Node.js stacks"},
	},
	{
		Name:           "pnpm",
		Category:       "JavaScript / TypeScript",
		Binary:         "pnpm",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `(\d+\.\d+\.\d+)`,
		MinVersion:     "8.0.0",
		InstallHelp:    "Install via 'npm install -g pnpm' or 'corepack enable'",
		UnlockedStacks: []string{"Fast Node.js package manager"},
	},
	{
		Name:           "bun",
		Category:       "JavaScript / TypeScript",
		Binary:         "bun",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `(\d+\.\d+\.\d+)`,
		MinVersion:     "1.0.0",
		InstallHelp:    "curl -fsSL https://bun.sh/install | bash",
		UnlockedStacks: []string{"High-performance JS/TS runtime"},
	},

	// Python Ecosystem
	{
		Name:           "Python 3",
		Category:       "Python",
		Binary:         "python3",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `Python (\d+\.\d+\.\d+)`,
		MinVersion:     "3.10.0",
		InstallHelp:    "Install via pyenv or package manager (apt/brew/dnf install python3)",
		UnlockedStacks: []string{"fastapi", "django"},
	},
	{
		Name:           "pip",
		Category:       "Python",
		Binary:         "pip3",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `pip (\d+\.\d+)`,
		MinVersion:     "22.0",
		InstallHelp:    "Install via 'python3 -m ensurepip' or 'apt install python3-pip'",
		UnlockedStacks: []string{"Python packages"},
	},

	// PHP Ecosystem
	{
		Name:           "PHP",
		Category:       "PHP",
		Binary:         "php",
		VersionArgs:    []string{"-v"},
		VersionRegex:   `PHP (\d+\.\d+\.\d+)`,
		MinVersion:     "8.2.0",
		InstallHelp:    "Install via 'apt install php php-cli' or 'brew install php'",
		UnlockedStacks: []string{"laravel"},
	},
	{
		Name:           "Composer",
		Category:       "PHP",
		Binary:         "composer",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `Composer (?:version )?(\d+\.\d+\.\d+)`,
		MinVersion:     "2.2.0",
		InstallHelp:    "curl -sS https://getcomposer.org/installer | php && sudo mv composer.phar /usr/local/bin/composer",
		UnlockedStacks: []string{"laravel"},
	},

	// Go Ecosystem
	{
		Name:           "Go",
		Category:       "Go",
		Binary:         "go",
		VersionArgs:    []string{"version"},
		VersionRegex:   `go version go(\d+\.\d+\.?\d*)`,
		MinVersion:     "1.21.0",
		InstallHelp:    "Install from https://go.dev/dl/ or 'brew install go'",
		UnlockedStacks: []string{"fiber", "gin"},
	},

	// Rust Ecosystem
	{
		Name:           "Rust / Cargo",
		Category:       "Rust",
		Binary:         "cargo",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `cargo (\d+\.\d+\.\d+)`,
		MinVersion:     "1.75.0",
		InstallHelp:    "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
		UnlockedStacks: []string{"axum"},
	},

	// Java / JVM Ecosystem
	{
		Name:           "Java JDK",
		Category:       "Java / JVM",
		Binary:         "java",
		VersionArgs:    []string{"-version"},
		VersionRegex:   `(?:version ")?(\d+(?:\.\d+)*)`,
		MinVersion:     "17.0.0",
		InstallHelp:    "Install OpenJDK 17 or 21 (apt install openjdk-21-jdk / brew install openjdk@21)",
		UnlockedStacks: []string{"springboot"},
	},

	// .NET Ecosystem
	{
		Name:           "dotnet SDK",
		Category:       ".NET",
		Binary:         "dotnet",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `(\d+\.\d+\.\d+)`,
		MinVersion:     "8.0.0",
		InstallHelp:    "Install .NET 8 SDK from https://dotnet.microsoft.com/download",
		UnlockedStacks: []string{"aspnet"},
	},

	// Mobile Ecosystem
	{
		Name:           "Flutter",
		Category:       "Mobile",
		Binary:         "flutter",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `Flutter (\d+\.\d+\.\d+)`,
		MinVersion:     "3.19.0",
		InstallHelp:    "Install Flutter SDK from https://docs.flutter.dev/get-started/install",
		UnlockedStacks: []string{"flutter"},
	},

	// DevOps & Containerization
	{
		Name:           "Docker Engine",
		Category:       "DevOps & Containers",
		Binary:         "docker",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `Docker version (\d+\.\d+\.\d+)`,
		MinVersion:     "20.10.0",
		InstallHelp:    "Install Docker Desktop or Docker Engine from https://docker.com",
		UnlockedStacks: []string{"Containerized builds & all databases"},
	},
	{
		Name:           "Git",
		Category:       "Version Control",
		Binary:         "git",
		VersionArgs:    []string{"--version"},
		VersionRegex:   `git version (\d+\.\d+\.\d+)`,
		MinVersion:     "2.20.0",
		InstallHelp:    "Install git via package manager (apt/brew/dnf install git)",
		UnlockedStacks: []string{"All project repositories"},
	},
}

// DiagnoseSystem runs a full audit of all toolchains
func DiagnoseSystem() []ToolDiagnostic {
	results := make([]ToolDiagnostic, 0, len(AuditedTools))

	for _, spec := range AuditedTools {
		diag := checkTool(spec)
		results = append(results, diag)
	}

	return results
}

func checkTool(spec ToolSpec) ToolDiagnostic {
	path, err := exec.LookPath(spec.Binary)
	if err != nil {
		return ToolDiagnostic{
			Name:           spec.Name,
			Category:       spec.Category,
			Binary:         spec.Binary,
			Status:         StatusMissing,
			MinVersion:     spec.MinVersion,
			InstallHelp:    spec.InstallHelp,
			Provisionable:  IsProvisionable(spec.Binary),
			UnlockedStacks: spec.UnlockedStacks,
		}
	}

	cmd := exec.Command(path, spec.VersionArgs...)
	out, _ := cmd.CombinedOutput()
	versionStr := "installed"
	if len(out) > 0 {
		re := regexp.MustCompile(spec.VersionRegex)
		match := re.FindStringSubmatch(string(out))
		if len(match) > 1 {
			versionStr = strings.TrimSpace(match[1])
		}
	}

	return ToolDiagnostic{
		Name:           spec.Name,
		Category:       spec.Category,
		Binary:         spec.Binary,
		Status:         StatusOK,
		Version:        versionStr,
		Path:           path,
		MinVersion:     spec.MinVersion,
		Provisionable:  IsProvisionable(spec.Binary),
		UnlockedStacks: spec.UnlockedStacks,
	}
}

// DiagnoseSystemJSON returns the diagnostic results as formatted JSON
func DiagnoseSystemJSON() (string, error) {
	diagnostics := DiagnoseSystem()
	b, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
