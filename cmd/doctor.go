package cmd

import (
	"github.com/opeteer/ace/internal/config"
	"github.com/opeteer/ace/internal/doctor"
	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var (
	flagDoctorJSON bool
	flagDoctorFix  bool
	flagDoctorAll  bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [framework | path]",
	Short: "Audit host environment or current project readiness",
	Long: `Audit your environment for required compilers, interpreters, package
managers, and container engines.

When run inside a project directory (or with a framework name like 'ace doctor laravel'),
ace doctor runs in scoped Project Doctor mode, checking only the requirements for that project.
Run 'ace doctor --all' (or outside a project) to audit all 19 universal host toolchains.

Examples:
  # Inside a project: audits only that project's requirements (e.g. PHP, Composer, .env, DB):
  ace doctor

  # Audit all 19 universal host toolchains across the entire operating system:
  ace doctor --all

  # Audit requirements for a specific framework:
  ace doctor laravel
  ace doctor next

  # Audit a project in a specific directory:
  ace doctor ./my-app

  # Output as machine-readable JSON:
  ace doctor --json

  # Audit and automatically download/install missing tools:
  ace doctor --fix`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !flagDoctorJSON {
			ui.PrintBanner(Version)
		}

		// 1. Explicit target passed as argument (framework name or project directory)
		if len(args) > 0 {
			target := args[0]

			// Check if target is a known framework ID
			if _, exists := config.FindFramework(target); exists {
				report, err := doctor.DiagnoseFramework(target)
				if err != nil {
					return err
				}
				return doctor.PrintProjectReport(report, flagDoctorJSON)
			}

			// Otherwise, treat as project path
			report, err := doctor.DiagnoseProject(target)
			if err != nil {
				return err
			}
			return doctor.PrintProjectReport(report, flagDoctorJSON)
		}

		// 2. If not forcing global audit (--all / --system), check if in a project directory
		if !flagDoctorAll {
			if _, ok := doctor.DetectProject("."); ok {
				report, err := doctor.DiagnoseProject(".")
				if err == nil {
					return doctor.PrintProjectReport(report, flagDoctorJSON)
				}
			}
		}

		// 3. Global system audit
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
	doctorCmd.Flags().BoolVar(&flagDoctorAll, "all", false, "Audit all 19 host toolchains across the entire operating system")
	doctorCmd.Flags().BoolVar(&flagDoctorAll, "system", false, "Alias for --all")
	rootCmd.AddCommand(doctorCmd)
}

