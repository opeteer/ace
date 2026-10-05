package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginTop(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA"))

	idStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D7D7"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all supported frameworks, databases, and add-ons",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintBanner(Version)

		fmt.Println(titleStyle.Render("📦 Supported Frameworks (19 Options)"))
		fmt.Printf("  %-12s %-24s %-12s %-25s %s\n",
			headerStyle.Render("ID"),
			headerStyle.Render("Framework"),
			headerStyle.Render("Language"),
			headerStyle.Render("Category"),
			headerStyle.Render("Description"))
		fmt.Println("  --------------------------------------------------------------------------------------------------")

		for _, f := range config.SupportedFrameworks {
			fmt.Printf("  %-12s %-24s %-12s %-25s %s\n",
				idStyle.Render(f.ID),
				f.Name,
				f.Language,
				string(f.Category),
				descStyle.Render(f.Description),
			)
		}

		fmt.Println()
		fmt.Println(titleStyle.Render("🗄 Supported Databases (10 Engines)"))
		fmt.Printf("  %-12s %-20s %-24s %-12s %s\n",
			headerStyle.Render("ID"),
			headerStyle.Render("Engine"),
			headerStyle.Render("Paradigm"),
			headerStyle.Render("Default Port"),
			headerStyle.Render("Docker Image"))
		fmt.Println("  --------------------------------------------------------------------------------------------------")

		for _, d := range config.SupportedDatabases {
			portStr := fmt.Sprintf("%d", d.DefaultPort)
			if d.DefaultPort == 0 {
				portStr = "embedded"
			}
			img := d.DockerImage
			if img == "" {
				img = "n/a"
			}
			fmt.Printf("  %-12s %-20s %-24s %-12s %s\n",
				idStyle.Render(d.ID),
				d.Name,
				d.Paradigm,
				portStr,
				descStyle.Render(img),
			)
		}

		fmt.Println()
		fmt.Println(titleStyle.Render("🛠 Supported Add-ons & Integrations"))
		fmt.Println("  DevOps:      --docker          (Generates multi-stage Dockerfile + docker-compose.yml)")
		fmt.Println("  Proxies:     --proxy nginx     (Nginx reverse proxy / FastCGI)")
		fmt.Println("               --proxy caddy     (Caddy reverse proxy with automatic HTTPS)")
		fmt.Println("  CI/CD:       --ci github       (GitHub Actions workflows)")
		fmt.Println("               --ci gitlab       (GitLab CI pipeline)")
		fmt.Println("  Protocols:   --protocol rest   (REST API routes)")
		fmt.Println("               --protocol grpc   (gRPC Protobuf services)")
		fmt.Println("               --protocol graphql(GraphQL server schemas)")
		fmt.Println("  Git:         --no-git          (Skip git init and initial commit)")
		fmt.Println()
	},
}
