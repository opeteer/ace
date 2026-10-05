package cmd

import (
	"github.com/opeteer/ace/internal/ui"
	"github.com/spf13/cobra"
)

var flagMono bool

var showtimeCmd = &cobra.Command{
	Use:   "showtime",
	Short: "Display the signature Ace spade ASCII art logo",
	Long:  "Display the high-definition signature Ace spade ASCII logo crafted from the official branding.",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintShowtime(flagMono)
	},
}

func init() {
	showtimeCmd.Flags().BoolVarP(&flagMono, "plain", "p", false, "Print monochrome ASCII art without ANSI colors")
	showtimeCmd.Flags().BoolVarP(&flagMono, "mono", "m", false, "Alias for --plain")
	rootCmd.AddCommand(showtimeCmd)
}
