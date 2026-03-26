package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunWithoutArgsShowsHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run(nil, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	assertRootHelp(t, stdout.String())
}

func TestRunWithUnknownFlagShowsHelpAndFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"--wat"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, `Error: unknown flag: --wat`) {
		t.Fatalf("expected unknown flag error, got %q", output)
	}

	assertRootHelp(t, output)
}

func TestRunWithUnknownSubcommandShowsHelpAndFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"scan"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, `Error: unknown command "scan" for "sonacli"`) {
		t.Fatalf("expected unknown command error, got %q", output)
	}

	assertRootHelp(t, output)
}

func assertRootHelp(t *testing.T, output string) {
	t.Helper()

	if !strings.Contains(output, "CLI for consuming SonarQube reports") {
		t.Fatalf("expected updated root summary, got %q", output)
	}

	if !strings.Contains(output, "Usage:\n  sonacli <command> <subcommand> [flags]\n") {
		t.Fatalf("expected custom usage block, got %q", output)
	}

	if !strings.Contains(output, "Available Commands:\n") {
		t.Fatalf("expected available commands section, got %q", output)
	}

	if !strings.Contains(output, "auth        Manage authentication settings to SonarQube") {
		t.Fatalf("expected updated auth command summary, got %q", output)
	}

	if !strings.Contains(output, "issue") {
		t.Fatalf("expected issue command to be listed, got %q", output)
	}

	if !strings.Contains(output, "skill") {
		t.Fatalf("expected skill command to be listed, got %q", output)
	}

	if !strings.Contains(output, "version") {
		t.Fatalf("expected version command to be listed, got %q", output)
	}

	if !strings.Contains(output, "Flags:\n  -h, --help") {
		t.Fatalf("expected help flag to be listed, got %q", output)
	}
}
