package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Category represents the framework high-level category
type Category string

const (
	CategoryFullstack Category = "Fullstack / SSR"
	CategoryBackend   Category = "Backend API"
	CategoryFrontend  Category = "Frontend SPA"
	CategoryMobile    Category = "Mobile (Cross-Platform)"
)

// FrameworkSpec defines metadata for a supported framework
type FrameworkSpec struct {
	ID          string
	Name        string
	Category    Category
	Language    string
	DefaultPort int
	Description string
}

// DatabaseSpec defines metadata for a supported database engine
type DatabaseSpec struct {
	ID          string
	Name        string
	Paradigm    string // Relational, Document, Key-Value, etc.
	DefaultPort int
	DockerImage string
	EnvPrefix   string
}

// ProxyType defines supported reverse proxies
type ProxyType string

const (
	ProxyNone    ProxyType = "none"
	ProxyNginx   ProxyType = "nginx"
	ProxyCaddy   ProxyType = "caddy"
	ProxyTraefik ProxyType = "traefik"
)

// ProtocolType defines supported API communication protocols
type ProtocolType string

const (
	ProtocolREST      ProtocolType = "rest"
	ProtocolGraphQL   ProtocolType = "graphql"
	ProtocolGRPC      ProtocolType = "grpc"
	ProtocolWebSocket ProtocolType = "websocket"
	ProtocolTRPC      ProtocolType = "trpc"
)

// CIType defines supported CI/CD platforms
type CIType string

const (
	CINone      CIType = "none"
	CIGitHub    CIType = "github"
	CIGitLab    CIType = "gitlab"
	CIBitbucket CIType = "bitbucket"
)

// ProjectConfig contains the full specification for a project to scaffold
type ProjectConfig struct {
	Name        string
	TargetPath  string
	Framework   FrameworkSpec
	Database    DatabaseSpec
	Redis       bool
	Docker      bool
	Proxy       ProxyType
	Protocol    ProtocolType
	CI          CIType
	NoGit       bool
	NoInstall   bool
	Interactive bool
}

// Validate checks if the configuration has valid required fields, sanitizes names, and links proxy to docker
func (c *ProjectConfig) Validate() error {
	rawName := strings.TrimSpace(c.Name)
	if rawName == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if c.Framework.ID == "" {
		return fmt.Errorf("framework must be specified")
	}
	if c.TargetPath == "" {
		c.TargetPath = filepath.Clean(rawName)
	}

	// Sanitize project name to prevent slashes and spaces in manifests/modules
	c.Name = Slugify(filepath.Base(rawName))

	// [BUG-06] If reverse proxy is requested, ensure Docker is enabled
	if c.Proxy != "" && c.Proxy != ProxyNone {
		c.Docker = true
	}

	return nil
}

// DBName returns a safe SQL-compliant database name with underscores (e.g. "my_app")
func (c *ProjectConfig) DBName() string {
	name := strings.ReplaceAll(c.Name, "-", "_")
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	res := strings.Trim(b.String(), "_")
	if res == "" {
		return "ace_db"
	}
	return res
}

// Slugify converts any string into a safe, URL and package-friendly slug (e.g. "my-cool-app")
func Slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' || r == '.' || r == '/' || r == '\\' || r == ':' {
			b.WriteRune('-')
		}
	}
	res := strings.Trim(b.String(), "-_")
	if res == "" {
		return "ace-app"
	}
	return res
}

