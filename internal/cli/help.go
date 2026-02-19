package cli

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	headerStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("34"))    // Greenish
	commandStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))               // Cyan
	flagStyle      = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("204")) // Pink
	separatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))              // Dark Gray
)

const coloredUsageTmpl = `{{Header "Usage:"}}
  {{if .Runnable}}{{Usage .UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{Command .CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{Header "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{Header "Examples:"}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{Header "Available Commands:"}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{Command (printf "%-11s" .Name)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{Header "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces | Flags}}{{end}}{{if .HasAvailableInheritedFlags}}

{{Header "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces | Flags}}{{end}}{{if .HasHelpSubCommands}}

{{Header "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{Command (printf "%-11s" .Name)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{Header "Use"}} {{Command (printf "%s [command] --help" .CommandPath)}} {{Header "for more information about a command."}}{{end}}
`

func colorizeHelp(cmd *cobra.Command) {
	cobra.AddTemplateFunc("Header", func(s string) string { return headerStyle.Render(s) })
	cobra.AddTemplateFunc("Command", func(s string) string { return commandStyle.Render(s) })

	// Flags function colorizes individual flag names and dimmed separators
	cobra.AddTemplateFunc("Flags", func(s string) string {
		reFlags := regexp.MustCompile(`(-\w|--[\w-]+)`)
		s = reFlags.ReplaceAllStringFunc(s, func(match string) string {
			return flagStyle.Render(match)
		})

		reSep := regexp.MustCompile(`, `)
		s = reSep.ReplaceAllString(s, separatorStyle.Render(", "))

		return s
	})

	// Usage function colorizes the top-level usage line including args
	cobra.AddTemplateFunc("Usage", func(s string) string {
		// Colorize <args> (required) - Yellow
		reArgs := regexp.MustCompile(`<[a-zA-Z0-9_-]+>`)
		s = reArgs.ReplaceAllStringFunc(s, func(match string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(match)
		})

		// Colorize [args] (optional/flags) - Dimmed
		// Use specific classes to avoid corrupting ANSI escape codes
		reFlags := regexp.MustCompile(`\[[a-zA-Z0-9_-]+\]`)
		s = reFlags.ReplaceAllStringFunc(s, func(match string) string {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(match)
		})

		// Colorize the command name (beginning of the line)
		reCmd := regexp.MustCompile(`^\w+`)
		s = reCmd.ReplaceAllStringFunc(s, func(match string) string {
			return commandStyle.Render(match)
		})

		return s
	})

	cmd.SetUsageTemplate(coloredUsageTmpl)
}
