package cmd

import (
	"github.com/opeteer/ace/internal/doctor"
	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var flagDoctorJSON bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Audit host environment and toolchain readiness",
	Long: `Audit your local machine for required compilers, interpreters, package
managers, and container engines (Node, Python, Go, PHP, Rust, Docker, etc.)
and display which frameworks can be run immediately.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !flagDoctorJSON {
			ui.PrintBanner(Version)
		}

		diagnostics := doctor.DiagnoseSystem()
		return doctor.PrintReport(diagnostics, flagDoctorJSON)
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&flagDoctorJSON, "json", false, "Output doctor diagnostics as machine-readable JSON")
	rootCmd.AddCommand(doctorCmd)
}
