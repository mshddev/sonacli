package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version can be overridden at build time with -ldflags "-X github.com/mshddev/sonacli/internal/cli.Version=vX.Y.Z".
var Version = "dev"

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the sonacli version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "sonacli %s\n", Version)
			return err
		},
	}

	applyCommandTemplates(cmd)

	return cmd
}
