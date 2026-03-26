package sonarqube

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateTokenAcceptsValidResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != validateTokenPath {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}

		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected accept header: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	if err := client.ValidateToken(context.Background(), "test-token"); err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
}

func TestValidateTokenRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "false payload",
			statusCode: http.StatusOK,
			body:       `{"valid":false}`,
		},
		{
			name:       "unauthorized status",
			statusCode: http.StatusUnauthorized,
			body:       `unauthorized`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			client := NewClient(server.URL, server.Client())

			err := client.ValidateToken(context.Background(), "test-token")
			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

func TestValidateTokenReturnsHelpfulErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		statusCode    int
		body          string
		errorContains string
	}{
		{
			name:          "invalid json",
			statusCode:    http.StatusOK,
			body:          `not-json`,
			errorContains: "decode auth validation response",
		},
		{
			name:          "unexpected status",
			statusCode:    http.StatusInternalServerError,
			body:          `boom`,
			errorContains: "unexpected SonarQube response status: 500 Internal Server Error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			client := NewClient(server.URL, server.Client())

			err := client.ValidateToken(context.Background(), "test-token")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.errorContains) {
				t.Fatalf("expected error containing %q, got %q", tc.errorContains, err.Error())
			}
		})
	}
}

func TestListIssuesReturnsResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != searchIssuesPath {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.URL.Query().Get("componentKeys"); got != "my-project" {
			t.Fatalf("unexpected componentKeys query: %q", got)
		}

		if got := r.URL.Query().Get("additionalFields"); got != "_all" {
			t.Fatalf("unexpected additionalFields query: %q", got)
		}

		if got := r.URL.Query().Get("p"); got != "2" {
			t.Fatalf("unexpected p query: %q", got)
		}

		if got := r.URL.Query().Get("ps"); got != "10" {
			t.Fatalf("unexpected ps query: %q", got)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-1"},{"key":"issue-2"}],"paging":{"pageIndex":2,"pageSize":10,"total":12}}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	result, err := client.ListIssues(context.Background(), "test-token", "my-project", IssueListOptions{Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("ListIssues returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	issues, ok := payload["issues"].([]any)
	if !ok {
		t.Fatalf("expected issues array, got %T", payload["issues"])
	}

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}
}

func TestListIssuesReturnsEmptyResponseForExistingProject(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case searchIssuesPath:
			if got := r.URL.Query().Get("componentKeys"); got != "my-project" {
				t.Fatalf("unexpected componentKeys query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"total":0}}`))
		case showComponentPath:
			if got := r.URL.Query().Get("component"); got != "my-project" {
				t.Fatalf("unexpected component query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	result, err := client.ListIssues(context.Background(), "test-token", "my-project", IssueListOptions{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListIssues returned error: %v", err)
	}

	if got := string(result); got != `{"issues":[],"paging":{"total":0}}` {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestListIssuesHandlesErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		statusCode    int
		body          string
		wantErr       error
		errorContains string
	}{
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `unauthorized`,
			wantErr:    ErrInvalidToken,
		},
		{
			name:          "unexpected status",
			statusCode:    http.StatusForbidden,
			body:          `forbidden`,
			errorContains: "unexpected SonarQube response status: 403 Forbidden",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			client := NewClient(server.URL, server.Client())

			_, err := client.ListIssues(context.Background(), "test-token", "my-project", IssueListOptions{Page: 1, PageSize: 20})
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.errorContains) {
				t.Fatalf("expected error containing %q, got %q", tc.errorContains, err.Error())
			}
		})
	}
}

func TestListIssuesReportsMissingProject(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case searchIssuesPath:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"total":0}}`))
		case showComponentPath:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":[{"msg":"Component key 'missing-project' not found"}]}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	_, err := client.ListIssues(context.Background(), "test-token", "missing-project", IssueListOptions{Page: 1, PageSize: 20})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("expected %v, got %v", ErrProjectNotFound, err)
	}
}

func TestListIssuesRejectsPagePastLastPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case searchIssuesPath:
			if got := r.URL.Query().Get("p"); got != "3" {
				t.Fatalf("unexpected p query: %q", got)
			}

			if got := r.URL.Query().Get("ps"); got != "20" {
				t.Fatalf("unexpected ps query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":3,"pageSize":20,"total":40}}`))
		case showComponentPath:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	_, err := client.ListIssues(context.Background(), "test-token", "my-project", IssueListOptions{Page: 3, PageSize: 20})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var pageErr *IssuePageOutOfRangeError
	if !errors.As(err, &pageErr) {
		t.Fatalf("expected IssuePageOutOfRangeError, got %v", err)
	}

	if pageErr.RequestedPage != 3 {
		t.Fatalf("unexpected requested page: %d", pageErr.RequestedPage)
	}

	if pageErr.LastPage != 2 {
		t.Fatalf("unexpected last page: %d", pageErr.LastPage)
	}
}

func TestGetIssueReturnsIssueJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != searchIssuesPath {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.URL.Query().Get("issues"); got != "issue-123" {
			t.Fatalf("unexpected issues query: %q", got)
		}

		if got := r.URL.Query().Get("additionalFields"); got != "_all" {
			t.Fatalf("unexpected additionalFields query: %q", got)
		}

		if got := r.URL.Query().Get("ps"); got != "1" {
			t.Fatalf("unexpected ps query: %q", got)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-123","message":"hello"}]}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	issue, err := client.GetIssue(context.Background(), "test-token", "issue-123")
	if err != nil {
		t.Fatalf("GetIssue returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(issue, &payload); err != nil {
		t.Fatalf("unmarshal issue: %v", err)
	}

	if got := payload["key"]; got != "issue-123" {
		t.Fatalf("unexpected issue key: %v", got)
	}
}

func TestGetIssueHandlesNotFoundAndInvalidToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
	}{
		{
			name:       "not found",
			statusCode: http.StatusOK,
			body:       `{"issues":[]}`,
			wantErr:    ErrIssueNotFound,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `unauthorized`,
			wantErr:    ErrInvalidToken,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			client := NewClient(server.URL, server.Client())

			_, err := client.GetIssue(context.Background(), "test-token", "issue-123")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestGetIssueReturnsHelpfulErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		statusCode    int
		body          string
		errorContains string
	}{
		{
			name:          "invalid json",
			statusCode:    http.StatusOK,
			body:          `not-json`,
			errorContains: "decode issue search response",
		},
		{
			name:          "unexpected status",
			statusCode:    http.StatusForbidden,
			body:          `forbidden`,
			errorContains: "unexpected SonarQube response status: 403 Forbidden",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			client := NewClient(server.URL, server.Client())

			_, err := client.GetIssue(context.Background(), "test-token", "issue-123")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.errorContains) {
				t.Fatalf("expected error containing %q, got %q", tc.errorContains, err.Error())
			}
		})
	}
}

func TestSearchProjectsReturnsResponseJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != searchProjectsPath {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.URL.Query().Get("p"); got != "1" {
			t.Fatalf("unexpected p query: %q", got)
		}

		if got := r.URL.Query().Get("ps"); got != "20" {
			t.Fatalf("unexpected ps query: %q", got)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":20,"total":2},"components":[{"key":"proj-1"},{"key":"proj-2"}]}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	result, err := client.SearchProjects(context.Background(), "test-token", 1, 20)
	if err != nil {
		t.Fatalf("SearchProjects returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	components, ok := payload["components"].([]any)
	if !ok {
		t.Fatalf("expected components array, got %T", payload["components"])
	}

	if len(components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(components))
	}
}

func TestSearchProjectsReturnsEmptyResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":20,"total":0},"components":[]}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	result, err := client.SearchProjects(context.Background(), "test-token", 1, 20)
	if err != nil {
		t.Fatalf("SearchProjects returned error: %v", err)
	}

	if got := string(result); got != `{"paging":{"pageIndex":1,"pageSize":20,"total":0},"components":[]}` {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestSearchProjectsHandlesErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		statusCode    int
		body          string
		wantErr       error
		errorContains string
	}{
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `unauthorized`,
			wantErr:    ErrInvalidToken,
		},
		{
			name:          "unexpected status",
			statusCode:    http.StatusForbidden,
			body:          `forbidden`,
			errorContains: "unexpected SonarQube response status: 403 Forbidden",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)

			client := NewClient(server.URL, server.Client())

			_, err := client.SearchProjects(context.Background(), "test-token", 1, 20)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.errorContains) {
				t.Fatalf("expected error containing %q, got %q", tc.errorContains, err.Error())
			}
		})
	}
}

func TestSearchProjectsRejectsPagePastLastPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("p"); got != "3" {
			t.Fatalf("unexpected p query: %q", got)
		}

		if got := r.URL.Query().Get("ps"); got != "20" {
			t.Fatalf("unexpected ps query: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paging":{"pageIndex":3,"pageSize":20,"total":40},"components":[]}`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, server.Client())

	_, err := client.SearchProjects(context.Background(), "test-token", 3, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var pageErr *ProjectPageOutOfRangeError
	if !errors.As(err, &pageErr) {
		t.Fatalf("expected ProjectPageOutOfRangeError, got %v", err)
	}

	if pageErr.RequestedPage != 3 {
		t.Fatalf("unexpected requested page: %d", pageErr.RequestedPage)
	}

	if pageErr.LastPage != 2 {
		t.Fatalf("unexpected last page: %d", pageErr.LastPage)
	}
}
