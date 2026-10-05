package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/generator"
	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var (
	flagFramework string
	flagDatabase  string
	flagDocker    bool
	flagProxy     string
	flagNginx     bool
	flagCI        string
	flagProtocol  string
	flagNoGit     bool
	flagNoInstall bool
)

var createCmd = &cobra.Command{
	Use:           "create [project-name]",
	Aliases:       []string{"new", "init"},
	Short:         "Scaffold a new project",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Scaffold a new web or mobile project with framework, database, Docker,
Nginx, and CI/CD ready to run.

Examples:
  # Interactive wizard:
  ace new

  # Laravel + PostgreSQL + Docker + Nginx + GitHub Actions:
  ace new my-laravel -f laravel -d postgres --docker --nginx --ci github

  # FastAPI + PostgreSQL + Docker:
  ace new my-api -f fastapi -d postgres --docker

  # Next.js + PostgreSQL:
  ace new my-web -f next -d postgres

  # Go Fiber + PostgreSQL + Docker:
  ace new my-service -f fiber -d postgres --docker`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner(Version)

		var projectName string
		if len(args) > 0 {
			projectName = args[0]
		}

		// Shorthand --nginx flag handling
		if flagNginx {
			flagProxy = "nginx"
		}

		var cfg *config.ProjectConfig

		// If no framework flag is supplied, invoke the interactive wizard
		if strings.TrimSpace(flagFramework) == "" {
			var err error
			cfg, err = ui.RunWizard(projectName)
			if err != nil {
				return err
			}
			cfg.TargetPath = cfg.Name
			cfg.Name = slugify(filepath.Base(cfg.Name))
			if cfg.Proxy != "" && cfg.Proxy != config.ProxyNone {
				cfg.Docker = true
			}
			cfg.NoInstall = flagNoInstall
		} else {
			// Flag-driven / headless mode
			if projectName == "" {
				projectName = "my-ace-app"
			}

			targetPath := projectName
			cleanName := slugify(filepath.Base(projectName))

			fw, ok := config.FindFramework(flagFramework)
			if !ok {
				return fmt.Errorf("unknown framework '%s'. Run 'ace list' to see available options", flagFramework)
			}

			var db config.DatabaseSpec
			if strings.TrimSpace(flagDatabase) != "" {
				var dbOk bool
				db, dbOk = config.FindDatabase(flagDatabase)
				if !dbOk {
					return fmt.Errorf("unknown database '%s'. Run 'ace list' to see available options", flagDatabase)
				}
			}

			proxyType := config.ProxyType(strings.ToLower(flagProxy))
			if proxyType == "" {
				proxyType = config.ProxyNone
			}

			// If proxy is enabled, ensure docker is enabled so compose coordinates them
			enableDocker := flagDocker
			if proxyType != config.ProxyNone {
				enableDocker = true
			}

			ciType := config.CIType(strings.ToLower(flagCI))
			if ciType == "" {
				ciType = config.CINone
			}

			protocolType := config.ProtocolType(strings.ToLower(flagProtocol))
			if protocolType == "" {
				protocolType = config.ProtocolREST
			}

			cfg = &config.ProjectConfig{
				Name:        cleanName,
				TargetPath:  targetPath,
				Framework:   fw,
				Database:    db,
				Docker:      enableDocker,
				Proxy:       proxyType,
				Protocol:    protocolType,
				CI:          ciType,
				NoGit:       flagNoGit,
				NoInstall:   flagNoInstall,
				Interactive: false,
			}
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		fmt.Printf("🚀 Scaffolding %s with %s...\n\n", cfg.Name, cfg.Framework.Name)

		engine := generator.NewEngine(cfg)
		if err := engine.Execute(); err != nil {
			return err
		}

		ui.PrintSummary(cfg)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&flagFramework, "framework", "f", "", "Framework to scaffold (e.g. laravel, fastapi, next, fiber)")
	createCmd.Flags().StringVarP(&flagDatabase, "db", "d", "", "Database to configure (e.g. postgres, mysql, mongo, sqlite, redis)")
	createCmd.Flags().BoolVar(&flagDocker, "docker", false, "Generate Dockerfile and docker-compose.yml")
	createCmd.Flags().StringVar(&flagProxy, "proxy", "", "Reverse proxy gateway (nginx, caddy)")
	createCmd.Flags().BoolVar(&flagNginx, "nginx", false, "Shorthand to enable Nginx reverse proxy")
	createCmd.Flags().StringVar(&flagCI, "ci", "", "CI/CD pipeline to generate (github, gitlab)")
	createCmd.Flags().StringVar(&flagProtocol, "protocol", "rest", "API communication protocol (rest, graphql, grpc, websocket)")
	createCmd.Flags().BoolVar(&flagNoGit, "no-git", false, "Do not initialize a Git repository")
	createCmd.Flags().BoolVar(&flagNoInstall, "no-install", false, "Skip automated dependency installation and sync hooks")
}

func slugify(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' || r == '.' || r == '/' || r == '\\' {
			b.WriteRune('-')
		}
	}
	res := strings.Trim(b.String(), "-_")
	if res == "" {
		return "ace-app"
	}
	return res
}

