package cli

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mshddev/sonacli/internal/config"
)

func TestRunAuthWithoutSubcommandShowsHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"auth"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Manage authentication settings to SonarQube") {
		t.Fatalf("expected auth help summary, got %q", output)
	}

	if !strings.Contains(output, "Usage:\n  sonacli auth <command> [flags]\n") {
		t.Fatalf("expected auth usage line, got %q", output)
	}

	if !strings.Contains(output, "setup") {
		t.Fatalf("expected setup command to be listed, got %q", output)
	}

	if !strings.Contains(output, "status") {
		t.Fatalf("expected status command to be listed, got %q", output)
	}
}

func TestRunAuthRejectsUnknownSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"auth", "unknown"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	want := `Error: unknown command "unknown" for "sonacli auth"

Manage authentication settings to SonarQube

Usage:
  sonacli auth <command> [flags]

Available Commands:
  setup       Setup the SonarQube server URL and token for sonacli
  status      Show sonacli authentication status to the SonarQube server
Flags:
  -h, --help   help for auth
`
	if got := stderr.String(); got != want {
		t.Fatalf("unexpected stderr:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunAuthSetupSavesConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/authentication/validate" {
			t.Fatalf("unexpected request path: %q", r.URL.Path)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("unexpected authorization header: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":true}`))
	}))
	t.Cleanup(server.Close)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"auth",
		"setup",
		"--server-url", server.URL + "/",
		"--token", "test-token",
	}, &stdout, &stderr)

	configPath, err := config.Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	wantStdout := "Saved SonarQube authentication settings.\n" +
		"Config file: " + configPath + "\n" +
		"Server URL: " + server.URL + "\n" +
		"Next:\n" +
		"  sonacli auth status\n" +
		"  sonacli project list\n" +
		"  sonacli issue list <project-key>\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("unexpected stdout: %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}

	want := "server_url: \"" + server.URL + "\"\ntoken: \"test-token\"\n"
	if got := string(data); got != want {
		t.Fatalf("unexpected config contents:\nwant %q\ngot  %q", want, got)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}

	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("unexpected config file mode: want 600, got %o", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(configPath))
	if err != nil {
		t.Fatalf("stat config directory: %v", err)
	}

	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("unexpected config directory mode: want 700, got %o", got)
	}
}

func TestRunAuthSetupRejectsInvalidToken(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"valid":false}`))
	}))
	t.Cleanup(server.Close)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"auth",
		"setup",
		"-s", server.URL,
		"-t", "bad-token",
	}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: token is not valid for the SonarQube server") {
		t.Fatalf("expected invalid token error, got %q", output)
	}

	if !strings.Contains(output, "Setup the SonarQube server URL and token for sonacli") {
		t.Fatalf("expected updated auth setup summary, got %q", output)
	}

	if !strings.Contains(output, "Examples:\n  sonacli auth setup --server-url <server-url> --token <token>\n  sonacli auth status\n") {
		t.Fatalf("expected auth setup help output, got %q", output)
	}

	configPath, err := config.Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("expected config file not to exist, got err=%v", err)
	}
}

func TestRunAuthSetupSurfacesValidationServiceErrors(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"auth",
		"setup",
		"-s", server.URL,
		"-t", "test-token",
	}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: validate token with SonarQube: unexpected SonarQube response status: 500 Internal Server Error") {
		t.Fatalf("expected validation service error, got %q", output)
	}

	if !strings.Contains(output, "Setup the SonarQube server URL and token for sonacli") {
		t.Fatalf("expected updated auth setup summary, got %q", output)
	}

	if !strings.Contains(output, "Examples:\n  sonacli auth setup --server-url <server-url> --token <token>\n  sonacli auth status\n") {
		t.Fatalf("expected auth setup help output, got %q", output)
	}
}

func TestRunAuthSetupMissingRequiredFlagsShowsHelpAndFails(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"auth", "setup"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, `Error: required flag(s) "server-url", "token" not set`) {
		t.Fatalf("expected missing required flags error, got %q", output)
	}

	if !strings.Contains(output, "Setup the SonarQube server URL and token for sonacli") {
		t.Fatalf("expected updated auth setup summary, got %q", output)
	}

	if !strings.Contains(output, "Examples:\n  sonacli auth setup --server-url <server-url> --token <token>\n  sonacli auth status\n") {
		t.Fatalf("expected auth setup help output, got %q", output)
	}

	if !strings.Contains(output, "--server-url string") {
		t.Fatalf("expected server-url flag help, got %q", output)
	}

	configPath, err := config.Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("expected config file not to exist, got err=%v", err)
	}
}

func TestRunAuthSetupRejectsInvalidServerURL(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"auth",
		"setup",
		"-s", "127.0.0.1:9000",
		"-t", "test-token",
	}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: server-url must use http or https") {
		t.Fatalf("expected invalid server-url error, got %q", output)
	}

	if !strings.Contains(output, "Setup the SonarQube server URL and token for sonacli") {
		t.Fatalf("expected updated auth setup summary, got %q", output)
	}

	if !strings.Contains(output, "Examples:\n  sonacli auth setup --server-url <server-url> --token <token>\n  sonacli auth status\n") {
		t.Fatalf("expected auth setup help output, got %q", output)
	}
}

func TestRunAuthStatusReportsConfiguredAuth(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	if _, err := config.SaveAuthSetup("http://127.0.0.1:9000", "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	configPath, err := config.Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"auth", "status"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	wantStdout := "SonarQube authentication is configured.\n" +
		"Config file: " + configPath + "\n" +
		"Server URL: http://127.0.0.1:9000\n" +
		"Token: test-*****\n\n" +
		"Useful commands:\n" +
		"  sonacli project list\n" +
		"  sonacli issue list <project-key>\n" +
		"  sonacli issue show <issue-key-or-url>\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("unexpected stdout: %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunAuthStatusReportsMissingAuthWithoutFailing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configPath, err := config.Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"auth", "status"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	wantStdout := "SonarQube authentication is not configured.\n" +
		"Config file: " + configPath + "\n" +
		"Save credentials with:\n" +
		"  sonacli auth setup --server-url <url> --token <token>\n" +
		"Then verify with:\n" +
		"  sonacli auth status\n"
	if got := stdout.String(); got != wantStdout {
		t.Fatalf("unexpected stdout: %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunAuthStatusFailsForMalformedConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configPath, err := config.Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte("server_url: http://127.0.0.1:9000\ntoken: \"test-token\"\n"), 0o600); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"auth", "status"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, `Error: load auth config "`) {
		t.Fatalf("expected auth config path in error, got %q", output)
	}

	if !strings.Contains(output, `parse config file: line 1: decode "server_url" value`) {
		t.Fatalf("expected malformed config error, got %q", output)
	}

	if !strings.Contains(output, "Show sonacli authentication status to the SonarQube server") {
		t.Fatalf("expected updated auth status summary, got %q", output)
	}

	if !strings.Contains(output, "Examples:\nsonacli auth status\n  sonacli auth setup --server-url <server-url> --token <token>\n") {
		t.Fatalf("expected auth status help output, got %q", output)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "long token",
			token: "test-token",
			want:  "test-*****",
		},
		{
			name:  "exactly five characters",
			token: "abcde",
			want:  "abcde",
		},
		{
			name:  "short token",
			token: "abcd",
			want:  "abcd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskToken(tt.token); got != tt.want {
				t.Fatalf("unexpected masked token: want %q, got %q", tt.want, got)
			}
		})
	}
}
