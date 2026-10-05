package cmd

import (
	"github.com/opeteer/ace/internal/doctor"
	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var (
	flagDoctorJSON bool
	flagDoctorFix  bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Audit host environment and toolchain readiness",
	Long: `Audit your local machine for required compilers, interpreters, package
managers, and container engines (Node, Python, Go, PHP, Rust, Docker, etc.)
and display which frameworks can be run immediately.

Examples:
  # Run diagnostics audit:
  ace doctor

  # Output as JSON:
  ace doctor --json

  # Audit and automatically download/install missing tools:
  ace doctor --fix`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !flagDoctorJSON {
			ui.PrintBanner(Version)
		}

		diagnostics := doctor.DiagnoseSystem()
		if err := doctor.PrintReport(diagnostics, flagDoctorJSON); err != nil {
			return err
		}

		if flagDoctorFix {
			return doctor.InstallMissingTools()
		}
		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&flagDoctorJSON, "json", false, "Output doctor diagnostics as machine-readable JSON")
	doctorCmd.Flags().BoolVar(&flagDoctorFix, "fix", false, "Automatically download and install missing tools")
	doctorCmd.Flags().BoolVar(&flagDoctorFix, "install", false, "Alias for --fix")
	rootCmd.AddCommand(doctorCmd)
}
