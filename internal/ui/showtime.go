package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// SpadeLogoLines contains the high-definition half-block ASCII art for the Ace spade logo
var SpadeLogoLines = []string{
	"                     ▄█                     ",
	"                    ▄███▄                   ",
	"                   ██████                   ",
	"                 ▄███████                   ",
	"                ▄████▀ ▀█  █▄               ",
	"              ▄█████▀   ▀  ███▄             ",
	"            ▄█████▀        █████▄           ",
	"          ▄███████         ███████▄         ",
	"        ▄█████████         █████████▄       ",
	"      ▄█████▀█████         ████▀▀█████▄     ",
	"    ▄█████▀  █████         ████   ▀█████▄   ",
	"   ▄████▀    █████         ████     ▀████▄  ",
	"  █████▀     █████         ████             ",
	" █████       █████         ████▄            ",
	"▄█████████████████         █████████████    ",
	"██████████████████         █████████████▄   ",
	"████▀        █████         ████▀            ",
	"█████        █████         ████             ",
	" ████▄       █████  ▄▄█▄▄  ████        ▄▄▄▄▄",
	" ▀█████▄▄    ████████████  ████▄   ▄▄▄█████ ",
	"   ▀█████    ████████████  ██████████████▀  ",
	"     ▀▀▀█    ████▀▀ █████  ▀█████████▀▀     ",
	"                    █████                   ",
	"                   ███████                  ",
	"                  ████████▄                 ",
	"                 ███████████                ",
	"               ▄█████████████▄              ",
}

// Electric cyber-neon palette transitioning through the spade
var spadeGradient = []string{
	"#FFFFFF", "#F3E8FF", "#E9D5FF", "#D8B4FE",
	"#C084FC", "#A855F7", "#9333EA", "#7E22CE",
	"#6B21A8", "#7C3AED", "#8B5CF6", "#A78BFA",
	"#818CF8", "#6366F1", "#4F46E5", "#4338CA",
	"#3B82F6", "#2563EB", "#1D4ED8", "#0EA5E9",
	"#06B6D4", "#14B8A6", "#10B981", "#059669",
	"#047857", "#065F46", "#044E3A",
}

// PrintShowtime renders the styled Ace spade logo
func PrintShowtime(mono bool) {
	fmt.Println()
	taglineStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00F5D4"))
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A855F7"))
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))

	for i, line := range SpadeLogoLines {
		if mono {
			fmt.Println(line)
		} else {
			colorHex := spadeGradient[i%len(spadeGradient)]
			lineStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorHex))
			fmt.Println(lineStyle.Render(line))
		}
	}

	fmt.Println()
	fmt.Println("       " + titleStyle.Render("♠  A C E  ♠"))
	fmt.Println("   " + taglineStyle.Render("The Universal Scaffolding Engine"))
	fmt.Println("  " + subtitleStyle.Render("One tool to scaffold, containerize & wire everything."))
	fmt.Println()
}
