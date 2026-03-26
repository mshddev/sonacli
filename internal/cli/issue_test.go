package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mshddev/sonacli/internal/config"
)

func TestRunIssueWithoutSubcommandShowsHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Read SonarQube issues") {
		t.Fatalf("expected issue help summary, got %q", output)
	}

	if !strings.Contains(output, "Usage:\n  sonacli issue <command> [flags]\n") {
		t.Fatalf("expected issue usage line, got %q", output)
	}

	if !strings.Contains(output, "list") {
		t.Fatalf("expected list command to be listed, got %q", output)
	}

	if !strings.Contains(output, "show") {
		t.Fatalf("expected show command to be listed, got %q", output)
	}
}

func TestRunIssueRejectsUnknownSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "unknown"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	want := `Error: unknown command "unknown" for "sonacli issue"

Read SonarQube issues

Usage:
  sonacli issue <command> [flags]

Available Commands:
  list        List SonarQube issues for a project as JSON
  show        Show a SonarQube issue as JSON
Flags:
  -h, --help   help for issue
`
	if got := stderr.String(); got != want {
		t.Fatalf("unexpected stderr:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunIssueListPrintsCompactJSONByDefault(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/search" {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.URL.Query().Get("componentKeys"); got != "my-project" {
			t.Fatalf("unexpected componentKeys query: %q", got)
		}

		if got := r.URL.Query().Get("p"); got != "1" {
			t.Fatalf("unexpected p query: %q", got)
		}

		if got := r.URL.Query().Get("ps"); got != "20" {
			t.Fatalf("unexpected ps query: %q", got)
		}

		if got := r.URL.Query().Get("issueStatuses"); got != "OPEN,CONFIRMED" {
			t.Fatalf("unexpected issueStatuses query: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-1"}],"paging":{"pageIndex":1,"pageSize":20,"total":1}}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	want := "{\"issues\":[{\"key\":\"issue-1\"}],\"paging\":{\"pageIndex\":1,\"pageSize\":20,\"total\":1}}\n\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunIssueListPrintsPrettyJSONWithFlag(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-1"}],"paging":{"total":1}}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--pretty"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}

	if !strings.Contains(stdout.String(), "\n") {
		t.Fatal("expected pretty-printed JSON with newlines")
	}
}

func TestRunIssueListSupportsExplicitPaginationFlags(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("p"); got != "2" {
			t.Fatalf("unexpected p query: %q", got)
		}

		if got := r.URL.Query().Get("ps"); got != "5" {
			t.Fatalf("unexpected ps query: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-6"}],"paging":{"pageIndex":2,"pageSize":5,"total":6}}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--page", "2", "--page-size", "5"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	want := "{\"issues\":[{\"key\":\"issue-6\"}],\"paging\":{\"pageIndex\":2,\"pageSize\":5,\"total\":6}}\n\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunIssueListRejectsMissingAuthSetup(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: SonarQube authentication is not configured; run sonacli auth setup --server-url <url> --token <token>") {
		t.Fatalf("expected missing auth error, got %q", output)
	}

	if !strings.Contains(output, "Examples:\n  sonacli issue list <project-key>\n") {
		t.Fatalf("expected issue list examples in help output, got %q", output)
	}
}

func TestRunIssueListRejectsInvalidPage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--page", "0"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: page must be greater than 0") {
		t.Fatalf("expected invalid page error, got %q", output)
	}
}

func TestRunIssueListRejectsInvalidPageSize(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--page-size", "0"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: page size must be greater than 0") {
		t.Fatalf("expected invalid page size error, got %q", output)
	}
}

func TestRunIssueListRejectsUnknownProject(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"total":0}}`))
		case "/api/components/show":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":[{"msg":"Component key 'missing-project' not found"}]}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "missing-project"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: project not found: missing-project") {
		t.Fatalf("expected missing project error, got %q", output)
	}
}

func TestRunIssueListRejectsPagePastLastPage(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":2,"pageSize":1,"total":1}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--page", "2", "--page-size", "1"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: page 2 is out of range: last page is 1") {
		t.Fatalf("expected last page error, got %q", output)
	}
}

func TestRunIssueListRejectsMissingProjectKey(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: accepts 1 arg(s), received 0") {
		t.Fatalf("expected missing arg error, got %q", output)
	}
}

func TestRunIssueListSendsCustomStatusFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("issueStatuses"); got != "ACCEPTED,FIXED" {
				t.Fatalf("unexpected issueStatuses query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--status", "ACCEPTED,FIXED"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListSendsSeverityFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("impactSeverities"); got != "HIGH,BLOCKER" {
				t.Fatalf("unexpected impactSeverities query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--severity", "HIGH,BLOCKER"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListSendsQualitiesFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("impactSoftwareQualities"); got != "SECURITY,RELIABILITY" {
				t.Fatalf("unexpected impactSoftwareQualities query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--qualities", "SECURITY,RELIABILITY"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListOmitsSeverityAndQualitiesWhenEmpty(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("impactSeverities") {
			t.Fatal("impactSeverities should not be present when not specified")
		}

		if r.URL.Query().Has("impactSoftwareQualities") {
			t.Fatal("impactSoftwareQualities should not be present when not specified")
		}

		if r.URL.Query().Has("assigned") {
			t.Fatal("assigned should not be present when not specified")
		}

		if r.URL.Query().Has("assignees") {
			t.Fatal("assignees should not be present when not specified")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-1"}],"paging":{"pageIndex":1,"pageSize":20,"total":1}}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListSendsAssignedFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("assigned"); got != "true" {
				t.Fatalf("unexpected assigned query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--assigned", "true"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListSendsAssigneesFilter(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("assignees"); got != "admin,__me__" {
				t.Fatalf("unexpected assignees query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--assignees", "admin,__me__"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListMeFlagSetsAssignees(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("assignees"); got != "__me__" {
				t.Fatalf("unexpected assignees query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--me"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListExpandsMeAlias(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("assignees"); got != "admin,__me__" {
				t.Fatalf("unexpected assignees query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--assignees", "admin,me"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListExpandsMineAlias(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/search":
			if got := r.URL.Query().Get("assignees"); got != "__me__" {
				t.Fatalf("unexpected assignees query: %q", got)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issues":[],"paging":{"pageIndex":1,"pageSize":20,"total":0}}`))
		case "/api/components/show":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"component":{"key":"my-project","qualifier":"TRK"}}`))
		default:
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--assignees", "mine"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}
}

func TestRunIssueListRejectsMeWithAssignees(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "list", "my-project", "--me", "--assignees", "admin"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: --me and --assignees are mutually exclusive") {
		t.Fatalf("expected mutually exclusive error, got %q", output)
	}
}

func TestRunIssueShowPrintsCompactJSONByDefault(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/search" {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.URL.Query().Get("issues"); got != "issue-123" {
			t.Fatalf("unexpected issues query: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-123","message":"hello","line":7}]}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "show", "issue-123"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	want := "{\"key\":\"issue-123\",\"message\":\"hello\",\"line\":7}\n\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunIssueShowPrintsPrettyJSONWithFlag(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[{"key":"issue-123","message":"hello","line":7}]}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup("https://configured.example", "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "show", server.URL + "/project/issues?id=sample&issues=issue-123", "--pretty"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	want := "{\n  \"key\": \"issue-123\",\n  \"message\": \"hello\",\n  \"line\": 7\n}\n\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunIssueShowReportsServerMismatchForURL(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup("https://configured.example", "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "show", server.URL + "/project/issues?id=sample&issues=issue-123"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: saved authentication is configured for https://configured.example, but the issue URL points to "+server.URL) {
		t.Fatalf("expected mismatch error, got %q", output)
	}
}

func TestRunIssueShowRejectsMissingAuthSetup(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "show", "issue-123"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: SonarQube authentication is not configured; run sonacli auth setup --server-url <url> --token <token>") {
		t.Fatalf("expected missing auth error, got %q", output)
	}

	if !strings.Contains(output, "Examples:\n  sonacli issue show AX1234567890\n") {
		t.Fatalf("expected issue show help output, got %q", output)
	}
}

func TestRunIssueShowRejectsMissingIssue(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[]}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"issue", "show", "issue-404"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: issue not found: issue-404") {
		t.Fatalf("expected missing issue error, got %q", output)
	}
}

func TestResolveIssueReference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		wantKey       string
		wantServerURL string
		wantErr       string
	}{
		{
			name:    "raw issue key",
			input:   "issue-123",
			wantKey: "issue-123",
		},
		{
			name:          "url query",
			input:         "https://sonar.example/project/issues?id=sample&issues=issue-123",
			wantKey:       "issue-123",
			wantServerURL: "https://sonar.example",
		},
		{
			name:          "url fragment",
			input:         "https://sonar.example/project/issues#id=sample&open=issue-123",
			wantKey:       "issue-123",
			wantServerURL: "https://sonar.example",
		},
		{
			name:          "url with base path",
			input:         "https://sonar.example/sonarqube/project/issues?id=sample&open=issue-123",
			wantKey:       "issue-123",
			wantServerURL: "https://sonar.example/sonarqube",
		},
		{
			name:    "missing key",
			input:   "https://sonar.example/project/issues?id=sample",
			wantErr: "issue URL must include one of the following query parameters",
		},
		{
			name:    "multiple keys",
			input:   "https://sonar.example/project/issues?issues=one,two",
			wantErr: "must contain exactly one issue key",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveIssueReference(tc.input)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("resolveIssueReference returned error: %v", err)
			}

			if got.Key != tc.wantKey {
				t.Fatalf("unexpected issue key: want %q, got %q", tc.wantKey, got.Key)
			}

			if got.ServerURL != tc.wantServerURL {
				t.Fatalf("unexpected server URL: want %q, got %q", tc.wantServerURL, got.ServerURL)
			}
		})
	}
}

func TestExpandAssigneeAliases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "me alone", input: "me", want: "__me__"},
		{name: "mine alone", input: "mine", want: "__me__"},
		{name: "me in list", input: "admin,me", want: "admin,__me__"},
		{name: "mine in list", input: "admin,mine,usera", want: "admin,__me__,usera"},
		{name: "__me__ passthrough", input: "__me__", want: "__me__"},
		{name: "no aliases", input: "admin,usera", want: "admin,usera"},
		{name: "trimmed spaces", input: " admin , me ", want: "admin,__me__"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := expandAssigneeAliases(tc.input)
			if got != tc.want {
				t.Fatalf("expandAssigneeAliases(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFormatJSONProducesValidJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		pretty bool
		want   string
	}{
		{
			name:   "compact by default",
			pretty: false,
			want:   "{\"key\":\"issue-123\"}\n\n",
		},
		{
			name:   "pretty with flag",
			pretty: true,
			want:   "{\n  \"key\": \"issue-123\"\n}\n\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := formatJSON([]byte(`{"key":"issue-123"}`), tc.pretty)
			if err != nil {
				t.Fatalf("formatJSON returned error: %v", err)
			}

			if string(got) != tc.want {
				t.Fatalf("unexpected formatted JSON: want %q, got %q", tc.want, string(got))
			}

			var payload map[string]any
			if err := json.Unmarshal(got, &payload); err != nil {
				t.Fatalf("formatted JSON is invalid: %v", err)
			}
		})
	}
}
