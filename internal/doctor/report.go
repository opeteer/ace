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

	summaryContent.WriteString(fmt.Sprintf("\n%s %s",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("Pro-Tip:"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("Ace generates multi-stage Docker builds. Use --docker to run any stack even without local runtimes!"),
	))

	b.WriteString("\n" + summaryBox.Render(summaryContent.String()))

	return b.String()
}
