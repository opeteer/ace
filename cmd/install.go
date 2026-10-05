package cmd

import (
	"fmt"

	"github.com/opeteer/ace/internal/doctor"
	"github.com/opeteer/ace/internal/generator/installer"
	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var flagInstallMissing bool

var installCmd = &cobra.Command{
	Use:     "install [tool | path]",
	Aliases: []string{"i"},
	Short:   "Install dependencies for a project or download missing toolchains from doctor",
	Long: `Download and install missing toolchains audited by 'ace doctor' (e.g. Composer,
Bun, pnpm, Rust/Cargo, .NET), or inspect a project directory and install all
manifest dependencies and framework sync hooks.

Examples:
  # Download and install all missing toolchains from ace doctor:
  ace install --missing
  # or alias:
  ace install --tools

  # Download and install a specific toolchain:
  ace install composer
  ace install bun
  ace install pnpm
  ace install rust

  # Install project dependencies in the current directory:
  ace install
  # or shorthand:
  ace i

  # Install project dependencies in a specific folder:
  ace install ./my-project`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner(Version)

		// 1. If --missing / --tools flag is passed, install all missing tools from doctor
		if flagInstallMissing {
			return doctor.InstallMissingTools()
		}

		// 2. If an argument is provided, check if it's a tool name first
		if len(args) > 0 {
			target := args[0]
			if doctor.IsProvisionable(target) {
				return doctor.InstallSpecificTool(target)
			}

			// Otherwise, treat as a project path
			fmt.Printf("📦 Inspecting project in %s...\n\n", target)
			return installer.InstallDirectory(target)
		}

		// 3. No arguments: attempt to install project dependencies in current directory
		fmt.Printf("📦 Inspecting project in current directory...\n\n")
		err := installer.InstallDirectory(".")
		if err != nil {
			// If not in a project, inform user and suggest --missing
			fmt.Printf("  ! %v\n\n", err)
			fmt.Printf("💡 To download and install missing development tools instead, run:\n")
			fmt.Printf("   ace install --missing   (or 'ace doctor --fix')\n\n")
		}
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&flagInstallMissing, "missing", false, "Download and install all missing tools identified by ace doctor")
	installCmd.Flags().BoolVar(&flagInstallMissing, "tools", false, "Alias for --missing")
	rootCmd.AddCommand(installCmd)
}
