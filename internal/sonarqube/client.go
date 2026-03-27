package sonarqube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	validateTokenPath  = "/api/authentication/validate"
	searchIssuesPath   = "/api/issues/search"
	searchProjectsPath = "/api/projects/search"
	showComponentPath  = "/api/components/show"
	defaultTimeout     = 10 * time.Second
)

var ErrInvalidToken = errors.New("token is not valid")
var ErrIssueNotFound = errors.New("issue not found")
var ErrProjectNotFound = errors.New("project not found")

type IssuePageOutOfRangeError struct {
	RequestedPage int
	LastPage      int
}

func (e *IssuePageOutOfRangeError) Error() string {
	if e.LastPage == 0 {
		return fmt.Sprintf("page %d is out of range: there are no issues for this project", e.RequestedPage)
	}

	return fmt.Sprintf("page %d is out of range: last page is %d", e.RequestedPage, e.LastPage)
}

type ProjectPageOutOfRangeError struct {
	RequestedPage int
	LastPage      int
}

func (e *ProjectPageOutOfRangeError) Error() string {
	if e.LastPage == 0 {
		return fmt.Sprintf("page %d is out of range: there are no projects", e.RequestedPage)
	}

	return fmt.Sprintf("page %d is out of range: last page is %d", e.RequestedPage, e.LastPage)
}

type IssueListOptions struct {
	Page       int
	PageSize   int
	Statuses   string
	Severities string
	Qualities  string
	Assigned   string
	Assignees  string
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type issueSearchResponse struct {
	Issues []json.RawMessage `json:"issues"`
	Paging issueSearchPaging `json:"paging"`
}

type issueSearchPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) ValidateToken(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+validateTokenPath, nil)
	if err != nil {
		return fmt.Errorf("create auth validation request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send auth validation request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var payload struct {
			Valid bool `json:"valid"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return fmt.Errorf("decode auth validation response: %w", err)
		}

		if !payload.Valid {
			return ErrInvalidToken
		}

		return nil
	case http.StatusUnauthorized:
		return ErrInvalidToken
	default:
		return fmt.Errorf("unexpected SonarQube response status: %s", resp.Status)
	}
}

func (c *Client) ListIssues(ctx context.Context, token, projectKey string, opts IssueListOptions) (json.RawMessage, error) {
	query := url.Values{}
	query.Set("componentKeys", projectKey)
	query.Set("additionalFields", "_all")
	query.Set("p", fmt.Sprintf("%d", opts.Page))
	query.Set("ps", fmt.Sprintf("%d", opts.PageSize))

	if opts.Statuses != "" {
		query.Set("issueStatuses", opts.Statuses)
	}
	if opts.Severities != "" {
		query.Set("impactSeverities", opts.Severities)
	}
	if opts.Qualities != "" {
		query.Set("impactSoftwareQualities", opts.Qualities)
	}
	if opts.Assigned != "" {
		query.Set("assigned", opts.Assigned)
	}
	if opts.Assignees != "" {
		query.Set("assignees", opts.Assignees)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+searchIssuesPath+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create issue list request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send issue list request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read issue list response: %w", err)
		}

		var payload issueSearchResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode issue list response: %w", err)
		}

		// SonarQube returns 200 with an empty issues array for an unknown project key,
		// so we need a second lookup to distinguish "missing project" from "no issues".
		if len(payload.Issues) == 0 {
			if err := c.ensureProjectExists(ctx, token, projectKey); err != nil {
				return nil, err
			}

			effectivePageSize := payload.Paging.PageSize
			if effectivePageSize <= 0 {
				effectivePageSize = opts.PageSize
			}

			if pageOutOfRange(opts.Page, payload.Paging.Total, effectivePageSize) {
				return nil, &IssuePageOutOfRangeError{
					RequestedPage: opts.Page,
					LastPage:      lastIssuePage(payload.Paging.Total, effectivePageSize),
				}
			}
		}

		return json.RawMessage(body), nil
	case http.StatusUnauthorized:
		return nil, ErrInvalidToken
	default:
		return nil, fmt.Errorf("unexpected SonarQube response status: %s", resp.Status)
	}
}

func pageOutOfRange(page, total, pageSize int) bool {
	if page <= 1 {
		return false
	}

	lastPage := lastIssuePage(total, pageSize)
	if lastPage == 0 {
		return true
	}

	return page > lastPage
}

func lastIssuePage(total, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}

	return (total + pageSize - 1) / pageSize
}

func (c *Client) ensureProjectExists(ctx context.Context, token, projectKey string) error {
	query := url.Values{}
	query.Set("component", projectKey)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+showComponentPath+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("create project lookup request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send project lookup request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return ErrInvalidToken
	case http.StatusNotFound:
		return ErrProjectNotFound
	default:
		return fmt.Errorf("unexpected SonarQube response status: %s", resp.Status)
	}
}

func (c *Client) GetIssue(ctx context.Context, token, issueKey string) (json.RawMessage, error) {
	query := url.Values{}
	query.Set("issues", issueKey)
	query.Set("additionalFields", "_all")
	query.Set("ps", "1")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+searchIssuesPath+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create issue search request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send issue search request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read issue search response: %w", err)
		}

		var payload struct {
			Issues []json.RawMessage `json:"issues"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode issue search response: %w", err)
		}

		if len(payload.Issues) == 0 {
			return nil, ErrIssueNotFound
		}

		if len(payload.Issues) > 1 {
			return nil, fmt.Errorf("expected exactly one issue in search response, got %d", len(payload.Issues))
		}

		return payload.Issues[0], nil
	case http.StatusUnauthorized:
		return nil, ErrInvalidToken
	default:
		return nil, fmt.Errorf("unexpected SonarQube response status: %s", resp.Status)
	}
}

type projectSearchResponse struct {
	Components []json.RawMessage   `json:"components"`
	Paging     projectSearchPaging `json:"paging"`
}

type projectSearchPaging struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
}

func (c *Client) SearchProjects(ctx context.Context, token string, page, pageSize int) (json.RawMessage, error) {
	query := url.Values{}
	query.Set("p", fmt.Sprintf("%d", page))
	query.Set("ps", fmt.Sprintf("%d", pageSize))

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+searchProjectsPath+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create project search request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send project search request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read project search response: %w", err)
		}

		var payload projectSearchResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode project search response: %w", err)
		}

		if len(payload.Components) == 0 {
			effectivePageSize := payload.Paging.PageSize
			if effectivePageSize <= 0 {
				effectivePageSize = pageSize
			}

			if pageOutOfRange(page, payload.Paging.Total, effectivePageSize) {
				return nil, &ProjectPageOutOfRangeError{
					RequestedPage: page,
					LastPage:      lastIssuePage(payload.Paging.Total, effectivePageSize),
				}
			}
		}

		return json.RawMessage(body), nil
	case http.StatusUnauthorized:
		return nil, ErrInvalidToken
	default:
		return nil, fmt.Errorf("unexpected SonarQube response status: %s", resp.Status)
	}
}
