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
  {{if .Runnable}}{{Command .UseLine}}{{end}}{{if .HasAvailableSubCommands}}
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

	// Add Flags function to colorize flag names and separators
	cobra.AddTemplateFunc("Flags", func(s string) string {
		// Colorize flags
		reFlags := regexp.MustCompile(`(-\w|--[\w-]+)`)
		s = reFlags.ReplaceAllStringFunc(s, func(match string) string {
			return flagStyle.Render(match)
		})

		// Colorize separator (comma between flags)
		reSep := regexp.MustCompile(`, `)
		s = reSep.ReplaceAllString(s, separatorStyle.Render(", "))

		return s
	})

	cmd.SetUsageTemplate(coloredUsageTmpl)
}
