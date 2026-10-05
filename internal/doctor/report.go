package doctor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF87"))

	categoryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginTop(1)

	badgeOK = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF87")).
		Render("✔ OK     ")

	badgeMissing = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F56")).
			Render("✖ MISSING")

	toolNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Width(16)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7D7")).
			Width(14)

	stacksStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5A93C"))

	summaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginTop(1)
)

// PrintReport prints the diagnostics report either as human-readable UI or JSON
func PrintReport(diagnostics []ToolDiagnostic, asJSON bool) error {
	if asJSON {
		out, err := json.MarshalIndent(diagnostics, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}

	fmt.Println(RenderReport(diagnostics))
	return nil
}

// RenderReport formats the diagnostic results into a styled string
func RenderReport(diagnostics []ToolDiagnostic) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🩺 Ace Toolchain & Environment Readiness Doctor") + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("Auditing local runtimes, package managers, and container engines...") + "\n")

	// Group by category
	categoryOrder := []string{
		"JavaScript / TypeScript",
		"Python",
		"PHP",
		"Go",
		"Rust",
		"Java / JVM",
		".NET",
		"Mobile",
		"DevOps & Containers",
		"Version Control",
	}

	grouped := make(map[string][]ToolDiagnostic)
	okCount := 0
	missingCount := 0
	readyStacks := make(map[string]bool)

	for _, d := range diagnostics {
		grouped[d.Category] = append(grouped[d.Category], d)
		if d.Status == StatusOK {
			okCount++
			for _, st := range d.UnlockedStacks {
				if !strings.HasPrefix(st, "All") && !strings.Contains(st, " ") {
					readyStacks[st] = true
				}
			}
		} else {
			missingCount++
		}
	}

	for _, cat := range categoryOrder {
		items, exists := grouped[cat]
		if !exists || len(items) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("\n%s\n", categoryStyle.Render("── "+cat+" ──")))
		for _, item := range items {
			var statusBadge string
			var ver string
			if item.Status == StatusOK {
				statusBadge = badgeOK
				ver = item.Version
				if ver == "" || ver == "installed" {
					ver = "available"
				}
			} else {
				statusBadge = badgeMissing
				ver = "-"
			}

			stacks := strings.Join(item.UnlockedStacks, ", ")
			b.WriteString(fmt.Sprintf("  %s %s %s %s\n",
				statusBadge,
				toolNameStyle.Render(item.Name),
				versionStyle.Render(ver),
				stacksStyle.Render(stacks),
			))

			if item.Status == StatusMissing && item.InstallHelp != "" {
				b.WriteString(fmt.Sprintf("             └─ %s\n", helpStyle.Render("Install: "+item.InstallHelp)))
			}
		}
	}

	// Summary card
	var readyStackList []string
	for st := range readyStacks {
		readyStackList = append(readyStackList, st)
	}

	var summaryContent strings.Builder
	summaryContent.WriteString(fmt.Sprintf("%s  Ready: %s  |  Missing: %s\n\n",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Render("Audit Summary:"),
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87")).Render(fmt.Sprintf("%d", okCount)),
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F56")).Render(fmt.Sprintf("%d", missingCount)),
	))

	if len(readyStackList) > 0 {
		summaryContent.WriteString(fmt.Sprintf("%s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Render("Instant Native Stacks:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render(strings.Join(readyStackList, ", ")),
		))
	}

	// Count provisionable missing tools
	var provMissing []string
	for _, d := range diagnostics {
		if d.Status == StatusMissing && d.Provisionable {
			provMissing = append(provMissing, d.Name)
		}
	}

	if len(provMissing) > 0 {
		summaryContent.WriteString(fmt.Sprintf("\n%s %s\n",
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87")).Render("Auto-Fix:"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Render(fmt.Sprintf("Run 'ace install --missing' (or 'ace doctor --fix') to install %d missing tool(s) (%s).", len(provMissing), strings.Join(provMissing, ", "))),
		))
	}

	summaryContent.WriteString(fmt.Sprintf("\n%s %s",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("Pro-Tip:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Ace generates multi-stage Docker builds. Use --docker to run any stack even without local runtimes!"),
	))

	b.WriteString("\n" + summaryBox.Render(summaryContent.String()))

	return b.String()
}

// PrintProjectReport prints the scoped project diagnostic report as UI or JSON
func PrintProjectReport(report *ProjectDiagnosticReport, asJSON bool) error {
	if asJSON {
		out, err := report.ToJSON()
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	}

	fmt.Println(RenderProjectReport(report))
	return nil
}

// RenderProjectReport renders a scoped project diagnostic report
func RenderProjectReport(report *ProjectDiagnosticReport) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("🩺 Ace Project Doctor: %s (%s)", report.Framework, report.Language)) + "\n")
	if report.ProjectPath != "-" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(fmt.Sprintf("Auditing requirements for %s at %s...", report.ProjectName, report.ProjectPath)) + "\n")
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(fmt.Sprintf("Auditing host readiness for %s...", report.Framework)) + "\n")
	}

	// Group checks by category
	categories := []string{"Toolchain", "Dependencies", "Configuration"}
	grouped := make(map[string][]ProjectCheckItem)
	for _, c := range report.Checks {
		cat := c.Category
		if cat == "" {
			cat = "Toolchain"
		}
		grouped[cat] = append(grouped[cat], c)
	}

	for _, cat := range categories {
		items, exists := grouped[cat]
		if !exists || len(items) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("\n%s\n", categoryStyle.Render("── "+cat+" ──")))
		for _, item := range items {
			var statusBadge string
			if item.Status == StatusOK {
				statusBadge = badgeOK
			} else {
				statusBadge = badgeMissing
			}

			cur := item.Current
			if cur == "" {
				cur = "-"
			}

			b.WriteString(fmt.Sprintf("  %s %s %s %s\n",
				statusBadge,
				toolNameStyle.Width(20).Render(item.Name),
				versionStyle.Width(34).Render(cur),
				stacksStyle.Render(item.Detail),
			))

			if item.Status == StatusMissing && item.InstallHelp != "" {
				b.WriteString(fmt.Sprintf("             └─ %s\n", helpStyle.Render("Fix: "+item.InstallHelp)))
			}
		}
	}

	// Project Summary Box
	var summary strings.Builder
	if report.IsReady {
		summary.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF87")).Render("✔ Project Status: Ready to develop and run!\n\n"))
		if len(report.NextSteps) > 0 {
			summary.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Render("Next Steps:\n"))
			for _, step := range report.NextSteps {
				summary.WriteString(fmt.Sprintf("  %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7D7")).Render(step)))
			}
			summary.WriteString("\n")
		}
	} else {
		summary.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F56")).Render("✖ Project Status: Action required to run project\n\n"))
		summary.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Render("Resolve the missing checks above to run this project natively.\n\n"))
	}

	summary.WriteString(fmt.Sprintf("%s %s",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("💡 Host Audit:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Run 'ace doctor --all' (or outside a project) to audit all 19 universal toolchains."),
	))

	b.WriteString("\n" + summaryBox.Render(summary.String()))
	return b.String()
}

