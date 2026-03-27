package cli

import (
	"context"
	"fmt"

	"github.com/mshddev/sonacli/internal/selfupdate"
	"github.com/spf13/cobra"
)

type updateRunner interface {
	Update(ctx context.Context, requestedVersion string) (selfupdate.Result, error)
}

var newUpdateRunner = func() (updateRunner, error) {
	return selfupdate.NewFromEnvironment(Version)
}

func NewUpdateCmd() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update sonacli to the latest version",
		Long:  "Download a sonacli release archive for the current operating system and CPU, verify checksums.txt, and replace the current executable in place. By default the latest GitHub release is installed.",
		Example: `  sonacli update
  sonacli update --version v0.1.0-rc.3`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			updater, err := newUpdateRunner()
			if err != nil {
				return err
			}

			result, err := updater.Update(cmd.Context(), version)
			if err != nil {
				return err
			}

			if !result.Updated {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "sonacli %s is already installed.\nPath: %s\n", result.Version, result.ExecutablePath)
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Updated sonacli from %s to %s.\nPath: %s\n", result.PreviousVersion, result.Version, result.ExecutablePath)
			return err
		},
	}

	applyCommandTemplates(cmd)
	cmd.Flags().StringVar(&version, "version", "", "Install a specific release tag instead of the latest release")

	return cmd
}
