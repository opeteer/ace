package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	accentStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D7D7"))
)

const bannerArt = `
    ___   ______ ______
   /   | / ____// ____/
  / /| |/ /    / __/   
 / ___ / /___ / /___   
/_/  |_|\____//_____/  
`

// PrintBanner prints the Ace branding header
func PrintBanner(version string) {
	fmt.Println(brandStyle.Render(bannerArt))
	fmt.Println(accentStyle.Render(" Ace CLI ") + subtleStyle.Render(fmt.Sprintf("v%s — Universal Automated Project Scaffolding", version)))
	fmt.Println(subtleStyle.Render(" One tool to scaffold, containerize, and wire everything."))
	fmt.Println()
}
