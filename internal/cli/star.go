package cli

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const repoURL = "https://github.com/mshddev/sonacli"

type browserOpener interface {
	Open(ctx context.Context, url string) error
}

type systemBrowserOpener struct {
	command string
}

var newStarOpener = func() (browserOpener, error) {
	return newSystemBrowserOpener(runtime.GOOS)
}

func NewStarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "star",
		Short:   "Star the sonacli GitHub repository",
		Long:    "Open the sonacli GitHub repository in your default browser so you can star it manually.",
		Example: `  sonacli star`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opener, err := newStarOpener()
			if err != nil {
				return err
			}

			if err := opener.Open(cmd.Context(), repoURL); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Opened %s in your browser.\n", repoURL)
			return err
		},
	}

	applyCommandTemplates(cmd)

	return cmd
}

func newSystemBrowserOpener(goos string) (*systemBrowserOpener, error) {
	switch goos {
	case "darwin":
		return &systemBrowserOpener{command: "open"}, nil
	case "linux":
		return &systemBrowserOpener{command: "xdg-open"}, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", goos)
	}
}

func (o *systemBrowserOpener) Open(ctx context.Context, url string) error {
	output, err := exec.CommandContext(ctx, o.command, url).CombinedOutput()
	if err == nil {
		return nil
	}

	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("open %s with %s: %w", url, o.command, err)
	}

	return fmt.Errorf("open %s with %s: %w: %s", url, o.command, err, message)
}
