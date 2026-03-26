package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/mshddev/sonacli/internal/config"
	"github.com/mshddev/sonacli/internal/sonarqube"
	"github.com/spf13/cobra"
)

func NewIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Read SonarQube issues",
		Args:  rejectUnknownSubcommands,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	applyCommandTemplates(cmd)
	cmd.AddCommand(NewIssueListCmd(), NewIssueShowCmd())

	return cmd
}

func expandAssigneeAliases(raw string) string {
	parts := strings.Split(raw, ",")
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		switch trimmed {
		case "me", "mine":
			parts[i] = "__me__"
		default:
			parts[i] = trimmed
		}
	}
	return strings.Join(parts, ",")
}

func NewIssueListCmd() *cobra.Command {
	var pretty bool
	var page int
	var pageSize int
	var statuses string
	var severities string
	var qualities string
	var assigned string
	var assignees string
	var me bool

	cmd := &cobra.Command{
		Use:   "list <project-key>",
		Short: "List SonarQube issues for a project as JSON",
		Long:  "Return the raw JSON payload from SonarQube /api/issues/search for one project. The default status filter is OPEN,CONFIRMED. Use --pretty for readable JSON.",
		Example: `  sonacli issue list <project-key>
  sonacli issue list <project-key> --pretty
  sonacli issue list <project-key> --status ACCEPTED,FIXED --page 2 --page-size 10
  sonacli issue list <project-key> --me`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectKey := strings.TrimSpace(args[0])
			if projectKey == "" {
				return errors.New("project key must not be empty")
			}

			if page < 1 {
				return errors.New("page must be greater than 0")
			}

			if pageSize < 1 {
				return errors.New("page size must be greater than 0")
			}

			if me && assignees != "" {
				return errors.New("--me and --assignees are mutually exclusive")
			}

			if me {
				assignees = "__me__"
			} else if assignees != "" {
				assignees = expandAssigneeAliases(assignees)
			}

			setup, err := config.LoadAuthSetup()
			if err != nil {
				if errors.Is(err, config.ErrAuthSetupNotFound) {
					return errors.New("SonarQube authentication is not configured; run sonacli auth setup --server-url <url> --token <token>")
				}

				return fmt.Errorf("load auth config: %w", err)
			}

			client := sonarqube.NewClient(setup.ServerURL, nil)
			issuesJSON, err := client.ListIssues(cmd.Context(), setup.Token, projectKey, sonarqube.IssueListOptions{
				Page:       page,
				PageSize:   pageSize,
				Statuses:   statuses,
				Severities: severities,
				Qualities:  qualities,
				Assigned:   assigned,
				Assignees:  assignees,
			})
			if err != nil {
				if errors.Is(err, sonarqube.ErrInvalidToken) {
					return errors.New("saved token is not valid for the SonarQube server")
				}
				if errors.Is(err, sonarqube.ErrProjectNotFound) {
					return fmt.Errorf("project not found: %s", projectKey)
				}

				var pageErr *sonarqube.IssuePageOutOfRangeError
				if errors.As(err, &pageErr) {
					return pageErr
				}

				return fmt.Errorf("fetch issues from SonarQube: %w", err)
			}

			formatted, err := formatJSON(issuesJSON, pretty)
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
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "Number of issues per page")
	cmd.Flags().StringVarP(&statuses, "status", "s", "OPEN,CONFIRMED", "Comma-separated issue statuses (OPEN,CONFIRMED,FALSE_POSITIVE,ACCEPTED,FIXED)")
	cmd.Flags().StringVarP(&severities, "severity", "e", "", "Comma-separated impact severities (INFO,LOW,MEDIUM,HIGH,BLOCKER)")
	cmd.Flags().StringVarP(&qualities, "qualities", "q", "", "Comma-separated software qualities (MAINTAINABILITY,RELIABILITY,SECURITY)")
	cmd.Flags().StringVarP(&assigned, "assigned", "a", "", "Filter assigned or unassigned issues (true, false)")
	cmd.Flags().StringVarP(&assignees, "assignees", "i", "", "Comma-separated list of assignee logins (use __me__ for current user)")
	cmd.Flags().BoolVarP(&me, "me", "m", false, "Shorthand for --assignees __me__")

	return cmd
}

func NewIssueShowCmd() *cobra.Command {
	var pretty bool

	cmd := &cobra.Command{
		Use:   "show <issue-id-or-url>",
		Short: "Show a SonarQube issue as JSON",
		Long:  "Return one SonarQube issue object as JSON. The argument may be a plain issue key or a SonarQube browser URL containing ?issue=, ?issues=, or ?open=.",
		Example: `  sonacli issue show AX1234567890
  sonacli issue show 'https://sonarqube.example.com/project/issues?id=my-project&issues=AX1234567890'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := resolveIssueReference(args[0])
			if err != nil {
				return err
			}

			setup, err := config.LoadAuthSetup()
			if err != nil {
				if errors.Is(err, config.ErrAuthSetupNotFound) {
					if ref.ServerURL != "" {
						return fmt.Errorf(
							"SonarQube authentication is not configured for %s; run sonacli auth setup --server-url %s --token <token>",
							ref.ServerURL,
							ref.ServerURL,
						)
					}

					return errors.New("SonarQube authentication is not configured; run sonacli auth setup --server-url <url> --token <token>")
				}

				return fmt.Errorf("load auth config: %w", err)
			}

			serverURL := setup.ServerURL
			configuredServerURL, err := normalizeServerURL(setup.ServerURL)
			if err == nil && ref.ServerURL != "" {
				serverURL = ref.ServerURL
			}

			client := sonarqube.NewClient(serverURL, nil)
			issueJSON, err := client.GetIssue(cmd.Context(), setup.Token, ref.Key)
			if err != nil {
				switch {
				case errors.Is(err, sonarqube.ErrInvalidToken):
					if ref.ServerURL != "" && configuredServerURL != "" && configuredServerURL != ref.ServerURL {
						return fmt.Errorf(
							"saved authentication is configured for %s, but the issue URL points to %s; run sonacli auth setup --server-url %s --token <token>",
							configuredServerURL,
							ref.ServerURL,
							ref.ServerURL,
						)
					}

					return errors.New("saved token is not valid for the SonarQube server")
				case errors.Is(err, sonarqube.ErrIssueNotFound):
					return fmt.Errorf("issue not found: %s", ref.Key)
				default:
					return fmt.Errorf("fetch issue from SonarQube: %w", err)
				}
			}

			formatted, err := formatJSON(issueJSON, pretty)
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(formatted)
			return err
		},
	}

	applyCommandTemplates(cmd)
	cmd.Flags().BoolVar(&pretty, "pretty", false, "Print pretty-printed JSON")

	return cmd
}

type issueReference struct {
	Key       string
	ServerURL string
}

func resolveIssueReference(raw string) (issueReference, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return issueReference{}, errors.New("issue reference must not be empty")
	}

	if !strings.Contains(trimmed, "://") {
		return issueReference{Key: trimmed}, nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return issueReference{}, fmt.Errorf("parse issue URL: %w", err)
	}

	issueKey, found, err := issueKeyFromValues(parsed.Query())
	if err != nil {
		return issueReference{}, err
	}
	if found {
		serverURL, err := issueServerURL(parsed)
		if err != nil {
			return issueReference{}, err
		}

		return issueReference{Key: issueKey, ServerURL: serverURL}, nil
	}

	if parsed.Fragment != "" {
		fragmentValues, fragmentErr := url.ParseQuery(strings.TrimPrefix(parsed.Fragment, "?"))
		if fragmentErr != nil {
			return issueReference{}, fmt.Errorf("parse issue URL fragment: %w", fragmentErr)
		}

		issueKey, found, err = issueKeyFromValues(fragmentValues)
		if err != nil {
			return issueReference{}, err
		}
		if found {
			serverURL, err := issueServerURL(parsed)
			if err != nil {
				return issueReference{}, err
			}

			return issueReference{Key: issueKey, ServerURL: serverURL}, nil
		}
	}

	return issueReference{}, errors.New("issue URL must include one of the following query parameters: issue, issues, open")
}

func issueServerURL(parsed *url.URL) (string, error) {
	basePath := strings.TrimRight(parsed.Path, "/")
	basePath = strings.TrimSuffix(basePath, "/project/issues")

	serverURL := &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   basePath,
	}

	normalized, err := normalizeServerURL(serverURL.String())
	if err != nil {
		return "", fmt.Errorf("derive server URL from issue URL: %w", err)
	}

	return normalized, nil
}

func issueKeyFromValues(values url.Values) (string, bool, error) {
	for _, name := range []string{"issue", "issues", "open"} {
		if !values.Has(name) {
			continue
		}

		issueKey, err := singleIssueKey(values.Get(name))
		if err != nil {
			return "", false, fmt.Errorf("%s query parameter: %w", name, err)
		}

		return issueKey, true, nil
	}

	return "", false, nil
}

func singleIssueKey(raw string) (string, error) {
	var issueKeys []string

	for _, part := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		issueKeys = append(issueKeys, trimmed)
	}

	switch len(issueKeys) {
	case 0:
		return "", errors.New("must contain exactly one issue key")
	case 1:
		return issueKeys[0], nil
	default:
		return "", errors.New("must contain exactly one issue key")
	}
}

func formatJSON(data []byte, pretty bool) ([]byte, error) {
	var output bytes.Buffer

	if pretty {
		if err := json.Indent(&output, data, "", "  "); err != nil {
			return nil, fmt.Errorf("format JSON: %w", err)
		}
	} else {
		if err := json.Compact(&output, data); err != nil {
			return nil, fmt.Errorf("format compact JSON: %w", err)
		}
	}

	output.WriteByte('\n')
	output.WriteByte('\n')

	return output.Bytes(), nil
}
