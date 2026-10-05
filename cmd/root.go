package cmd

import (
	"fmt"
	"os"

	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var Version = "0.1.0-poc"

var rootCmd = &cobra.Command{
	Use:     "ace",
	Version: Version,
	Short:   "Ace - Automated Scaffolding for Web & Mobile Frameworks",
	Long: `Ace is a high-performance CLI tool for automated scaffolding across all
popular web and mobile frameworks, with instant wiring for databases,
Docker, Nginx, protocols, and CI/CD pipelines.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintBanner(Version)
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
}
