package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const rootUsageTemplate = `Usage:
  sonacli <command> <subcommand> [flags]

{{if .HasAvailableSubCommands}}Available Commands:
{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "completion"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}`

const commandUsageTemplate = `Usage:
  {{if .HasAvailableSubCommands}}{{.CommandPath}} <command>{{if .HasAvailableLocalFlags}} [flags]{{end}}{{else}}{{.UseLine}}{{end}}

{{if .HasAvailableSubCommands}}Available Commands:
{{range .Commands}}{{if (and .IsAvailableCommand (ne .Name "completion"))}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}`

const helpTemplate = `{{if .Short}}{{.Short | trimTrailingWhitespaces}}{{end}}{{if .Long}}

{{.Long | trimTrailingWhitespaces}}{{end}}

{{.UsageString}}{{if .Example}}Examples:
{{.Example | trimTrailingWhitespaces}}
{{end}}`

func NewRootCmd(stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "sonacli",
		Short:         "CLI for consuming SonarQube reports",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          rejectUnknownSubcommands,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetUsageTemplate(rootUsageTemplate)
	cmd.SetHelpTemplate(helpTemplate)
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return err
	})
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.InitDefaultHelpFlag()
	cmd.AddCommand(NewAuthCmd(), NewIssueCmd(), NewProjectCmd(), NewSkillCmd(), NewVersionCmd())

	return cmd
}

func applyCommandTemplates(cmd *cobra.Command) {
	cmd.SetUsageTemplate(commandUsageTemplate)
	cmd.SetHelpTemplate(helpTemplate)
}

func rejectUnknownSubcommands(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
}

func Run(args []string, stdout, stderr io.Writer) int {
	cmd := NewRootCmd(stdout, stderr)
	cmd.SetArgs(args)

	executedCmd, err := cmd.ExecuteC()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n\n", err)

		helpCmd := executedCmd
		if helpCmd == nil {
			helpCmd = cmd
		}

		helpCmd.SetOut(stderr)

		if helpErr := helpCmd.Help(); helpErr != nil {
			fmt.Fprintf(stderr, "Error: %v\n", helpErr)
		}

		return 1
	}

	return 0
}
