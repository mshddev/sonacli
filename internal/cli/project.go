package cli

import (
	"errors"
	"fmt"

	"github.com/mshddev/sonacli/internal/config"
	"github.com/mshddev/sonacli/internal/sonarqube"
	"github.com/spf13/cobra"
)

func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Read SonarQube projects",
		Args:  rejectUnknownSubcommands,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	applyCommandTemplates(cmd)
	cmd.AddCommand(NewProjectListCmd())

	return cmd
}

func NewProjectListCmd() *cobra.Command {
	var pretty bool
	var page int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SonarQube projects as JSON",
		Long:  "Return the raw JSON payload from SonarQube /api/projects/search. Use --page and --page-size for pagination and --pretty for readable JSON.",
		Example: `  sonacli project list
  sonacli project list --pretty
  sonacli project list --page 2 --page-size 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if page < 1 {
				return errors.New("page must be greater than 0")
			}

			if pageSize < 1 {
				return errors.New("page size must be greater than 0")
			}

			setup, err := config.LoadAuthSetup()
			if err != nil {
				if errors.Is(err, config.ErrAuthSetupNotFound) {
					return errors.New("SonarQube authentication is not configured; run sonacli auth setup --server-url <url> --token <token>")
				}

				return fmt.Errorf("load auth config: %w", err)
			}

			client := sonarqube.NewClient(setup.ServerURL, nil)
			projectsJSON, err := client.SearchProjects(cmd.Context(), setup.Token, page, pageSize)
			if err != nil {
				if errors.Is(err, sonarqube.ErrInvalidToken) {
					return errors.New("saved token is not valid for the SonarQube server")
				}

				var pageErr *sonarqube.ProjectPageOutOfRangeError
				if errors.As(err, &pageErr) {
					return pageErr
				}

				return fmt.Errorf("fetch projects from SonarQube: %w", err)
			}

			formatted, err := formatJSON(projectsJSON, pretty)
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(formatted)
			return err
		},
	}

	applyCommandTemplates(cmd)
	cmd.Flags().BoolVar(&pretty, "pretty", false, "Print pretty-printed JSON")
	cmd.Flags().IntVar(&page, "page", 1, "Page number to request")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Number of projects per page")

	return cmd
}
