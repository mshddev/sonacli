package cli

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/mshddev/sonacli/internal/config"
	"github.com/mshddev/sonacli/internal/sonarqube"
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication settings to SonarQube",
		Args:  rejectUnknownSubcommands,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	applyCommandTemplates(cmd)
	cmd.AddCommand(NewAuthSetupCmd())
	cmd.AddCommand(NewAuthStatusCmd())

	return cmd
}

func NewAuthSetupCmd() *cobra.Command {
	var serverURL string
	var token string

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup the SonarQube server URL and token for sonacli",
		Example: `  sonacli auth setup --server-url <server-url> --token <token>
  sonacli auth status`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			normalizedServerURL, err := normalizeServerURL(serverURL)
			if err != nil {
				return err
			}

			trimmedToken := strings.TrimSpace(token)
			if trimmedToken == "" {
				return errors.New("token must not be empty")
			}

			client := sonarqube.NewClient(normalizedServerURL, nil)
			if err := client.ValidateToken(cmd.Context(), trimmedToken); err != nil {
				if errors.Is(err, sonarqube.ErrInvalidToken) {
					return errors.New("token is not valid for the SonarQube server")
				}

				return fmt.Errorf("validate token with SonarQube: %w", err)
			}

			configPath, err := config.SaveAuthSetup(normalizedServerURL, trimmedToken)
			if err != nil {
				return fmt.Errorf("save auth config: %w", err)
			}

			_, err = fmt.Fprintf(
				cmd.OutOrStdout(),
				"Saved SonarQube authentication settings.\nConfig file: %s\nServer URL: %s\nNext:\n  sonacli auth status\n  sonacli project list\n  sonacli issue list <project-key>\n",
				configPath,
				normalizedServerURL,
			)
			return err
		},
	}

	applyCommandTemplates(cmd)
	cmd.Flags().StringVarP(&serverURL, "server-url", "s", "", "SonarQube server URL")
	cmd.Flags().StringVarP(&token, "token", "t", "", "SonarQube user token")
	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}

func NewAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sonacli authentication status to the SonarQube server",
		Example: `sonacli auth status
  sonacli auth setup --server-url <server-url> --token <token>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.Path()
			if err != nil {
				return fmt.Errorf("resolve auth config path: %w", err)
			}

			setup, err := config.LoadAuthSetup()
			if err != nil {
				if errors.Is(err, config.ErrAuthSetupNotFound) {
					_, err = fmt.Fprintln(cmd.OutOrStdout(), "SonarQube authentication is not configured.")
					if err != nil {
						return err
					}

					_, err = fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n", configPath)
					if err != nil {
						return err
					}

					_, err = fmt.Fprintln(cmd.OutOrStdout(), "Save credentials with:")
					if err != nil {
						return err
					}

					_, err = fmt.Fprintln(cmd.OutOrStdout(), "  sonacli auth setup --server-url <url> --token <token>")
					if err != nil {
						return err
					}

					_, err = fmt.Fprintln(cmd.OutOrStdout(), "Then verify with:")
					if err != nil {
						return err
					}

					_, err = fmt.Fprintln(cmd.OutOrStdout(), "  sonacli auth status")
					return err
				}

				return fmt.Errorf("load auth config %q: %w", configPath, err)
			}

			_, err = fmt.Fprintf(
				cmd.OutOrStdout(),
				"SonarQube authentication is configured.\nConfig file: %s\nServer URL: %s\nToken: %s\n\nUseful commands:\n  sonacli project list\n  sonacli issue list <project-key>\n  sonacli issue show <issue-key-or-url>\n",
				configPath,
				setup.ServerURL,
				maskToken(setup.Token),
			)
			return err
		},
	}

	applyCommandTemplates(cmd)

	return cmd
}

func maskToken(token string) string {
	const visiblePrefix = 5

	runes := []rune(token)
	if len(runes) <= visiblePrefix {
		return token
	}

	return string(runes[:visiblePrefix]) + strings.Repeat("*", len(runes)-visiblePrefix)
}

func normalizeServerURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("server-url must not be empty")
	}

	if !strings.Contains(trimmed, "://") {
		return "", errors.New("server-url must use http or https")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse server-url: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("server-url must use http or https")
	}

	if parsed.Host == "" {
		return "", errors.New("server-url must include a host")
	}

	if parsed.User != nil {
		return "", errors.New("server-url must not include user info")
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("server-url must not include a query or fragment")
	}

	return strings.TrimRight(parsed.String(), "/"), nil
}
