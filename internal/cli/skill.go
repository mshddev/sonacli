package cli

import (
	"fmt"

	"github.com/mshddev/sonacli/internal/agentskill"
	"github.com/spf13/cobra"
)

func NewSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage agent skills for sonacli",
		Args:  rejectUnknownSubcommands,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	applyCommandTemplates(cmd)
	cmd.AddCommand(NewSkillInstallCmd(), NewSkillUninstallCmd())

	return cmd
}

func NewSkillInstallCmd() *cobra.Command {
	var includeCodex bool
	var includeClaude bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the sonacli skill for supported agents",
		Long:  "Install the managed sonacli skill into the detected agent skill directories. Pass --codex or --claude to skip PATH detection and target a specific agent explicitly.",
		Example: `  sonacli skill install
  sonacli skill install --codex
  sonacli skill install --claude`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := agentskill.Manager{}

			targets, err := manager.ResolveInstallTargets(includeCodex, includeClaude)
			if err != nil {
				return err
			}

			results, err := manager.Install(targets)
			if err != nil {
				return err
			}

			for _, result := range results {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Installed sonacli skill for %s.\nPath: %s\n", result.Target, result.Path); err != nil {
					return err
				}
			}

			return nil
		},
	}

	applyCommandTemplates(cmd)
	cmd.Flags().BoolVar(&includeCodex, "codex", false, "Install the skill for Codex")
	cmd.Flags().BoolVar(&includeClaude, "claude", false, "Install the skill for Claude Code")

	return cmd
}

func NewSkillUninstallCmd() *cobra.Command {
	var includeCodex bool
	var includeClaude bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the installed sonacli skill from supported agents",
		Long:  "Remove only skill directories managed by sonacli. Pass --codex or --claude to uninstall from a specific agent target.",
		Example: `  sonacli skill uninstall
  sonacli skill uninstall --codex
  sonacli skill uninstall --claude`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := agentskill.Manager{}

			targets := manager.ResolveUninstallTargets(includeCodex, includeClaude)
			results, err := manager.Uninstall(targets)
			if err != nil {
				return err
			}

			for _, result := range results {
				if !result.Removed {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "sonacli skill is not installed for %s.\nPath: %s\n", result.Target, result.Path); err != nil {
						return err
					}

					continue
				}

				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Removed sonacli skill for %s.\nPath: %s\n", result.Target, result.Path); err != nil {
					return err
				}
			}

			return nil
		},
	}

	applyCommandTemplates(cmd)
	cmd.Flags().BoolVar(&includeCodex, "codex", false, "Remove the skill from Codex")
	cmd.Flags().BoolVar(&includeClaude, "claude", false, "Remove the skill from Claude Code")

	return cmd
}
