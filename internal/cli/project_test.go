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

func TestRunProjectWithoutSubcommandShowsHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Read SonarQube projects") {
		t.Fatalf("expected project help summary, got %q", output)
	}

	if !strings.Contains(output, "Usage:\n  sonacli project <command> [flags]\n") {
		t.Fatalf("expected project usage line, got %q", output)
	}

	if !strings.Contains(output, "list") {
		t.Fatalf("expected list command to be listed, got %q", output)
	}
}

func TestRunProjectRejectsUnknownSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "unknown"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	want := `Error: unknown command "unknown" for "sonacli project"

Read SonarQube projects

Usage:
  sonacli project <command> [flags]

Available Commands:
  list        List SonarQube projects as JSON
Flags:
  -h, --help   help for project
`
	if got := stderr.String(); got != want {
		t.Fatalf("unexpected stderr:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunProjectListPrintsCompactJSONByDefault(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/search" {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.URL.Query().Get("p"); got != "1" {
			t.Fatalf("unexpected p query: %q", got)
		}

		if got := r.URL.Query().Get("ps"); got != "20" {
			t.Fatalf("unexpected ps query: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paging":{"pageIndex":1,"pageSize":20,"total":1},"components":[{"key":"my-project","name":"My Project"}]}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	want := "{\"paging\":{\"pageIndex\":1,\"pageSize\":20,\"total\":1},\"components\":[{\"key\":\"my-project\",\"name\":\"My Project\"}]}\n\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunProjectListPrintsPrettyJSONWithFlag(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paging":{"total":1},"components":[{"key":"my-project"}]}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list", "--pretty"}, &stdout, &stderr)

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

func TestRunProjectListSupportsExplicitPaginationFlags(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"paging":{"pageIndex":2,"pageSize":5,"total":6},"components":[{"key":"project-6"}]}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list", "--page", "2", "--page-size", "5"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	want := "{\"paging\":{\"pageIndex\":2,\"pageSize\":5,\"total\":6},\"components\":[{\"key\":\"project-6\"}]}\n\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunProjectListRejectsMissingAuthSetup(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list"}, &stdout, &stderr)

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

	if !strings.Contains(output, "Examples:\n  sonacli project list\n") {
		t.Fatalf("expected project list examples in help output, got %q", output)
	}
}

func TestRunProjectListRejectsInvalidPage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list", "--page", "0"}, &stdout, &stderr)

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

func TestRunProjectListRejectsInvalidPageSize(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list", "--page-size", "0"}, &stdout, &stderr)

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

func TestRunProjectListRejectsInvalidToken(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: saved token is not valid for the SonarQube server") {
		t.Fatalf("expected invalid token error, got %q", output)
	}
}

func TestRunProjectListRejectsPagePastLastPage(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paging":{"pageIndex":2,"pageSize":1,"total":1},"components":[]}`))
	}))
	t.Cleanup(server.Close)

	if _, err := config.SaveAuthSetup(server.URL, "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list", "--page", "2", "--page-size", "1"}, &stdout, &stderr)

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

func TestRunProjectListRejectsUnexpectedArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"project", "list", "extra-arg"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: unknown command") {
		t.Fatalf("expected unknown command error, got %q", output)
	}
}
