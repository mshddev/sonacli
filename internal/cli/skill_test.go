package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mshddev/sonacli/internal/agentskill"
)

func TestRunSkillWithoutSubcommandShowsHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Manage agent skills for sonacli") {
		t.Fatalf("expected skill help summary, got %q", output)
	}

	if !strings.Contains(output, "Usage:\n  sonacli skill <command> [flags]\n") {
		t.Fatalf("expected skill usage line, got %q", output)
	}

	if !strings.Contains(output, "install") {
		t.Fatalf("expected install command to be listed, got %q", output)
	}

	if !strings.Contains(output, "uninstall") {
		t.Fatalf("expected uninstall command to be listed, got %q", output)
	}
}

func TestRunSkillRejectsUnknownSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill", "unknown"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	want := `Error: unknown command "unknown" for "sonacli skill"

Manage agent skills for sonacli

Usage:
  sonacli skill <command> [flags]

Available Commands:
  install     Install the sonacli skill for supported agents
  uninstall   Remove the installed sonacli skill from supported agents
Flags:
  -h, --help   help for skill
`
	if got := stderr.String(); got != want {
		t.Fatalf("unexpected stderr:\nwant %q\ngot  %q", want, got)
	}
}

func TestRunSkillInstallDetectsTargetsAndInstallsManagedSkills(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	binDir := t.TempDir()
	writeCLIExecutable(t, binDir, "codex")
	writeCLIExecutable(t, binDir, "claude")
	t.Setenv("PATH", binDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill", "install"}, &stdout, &stderr)

	codexPath := filepath.Join(homeDir, ".codex", "skills", agentskill.SkillName)
	claudePath := filepath.Join(homeDir, ".claude", "skills", agentskill.SkillName)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	want := "Installed sonacli skill for codex.\nPath: " + codexPath + "\n" +
		"Installed sonacli skill for claude.\nPath: " + claudePath + "\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	assertPathExists(t, codexPath)
	assertPathExists(t, claudePath)
}

func TestRunSkillInstallWithExplicitTargetBypassesDetection(t *testing.T) {
	homeDir := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("PATH", t.TempDir())

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill", "install", "--codex"}, &stdout, &stderr)

	codexPath := filepath.Join(codexHome, "skills", agentskill.SkillName)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	want := "Installed sonacli skill for codex.\nPath: " + codexPath + "\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	assertPathExists(t, codexPath)
}

func TestRunSkillInstallRejectsWhenNothingDetected(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")
	t.Setenv("PATH", t.TempDir())

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill", "install"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: no supported agent CLI detected on PATH: codex, claude; pass --codex or --claude to install explicitly") {
		t.Fatalf("expected detection error, got %q", output)
	}

	if !strings.Contains(output, "Install the managed sonacli skill into the detected agent skill directories.") {
		t.Fatalf("expected updated skill install long help, got %q", output)
	}

	if !strings.Contains(output, "Examples:\n  sonacli skill install\n  sonacli skill install --codex\n  sonacli skill install --claude\n") {
		t.Fatalf("expected skill install help output, got %q", output)
	}
}

func TestRunSkillInstallRejectsUnmanagedExistingSkillDirectory(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	binDir := t.TempDir()
	writeCLIExecutable(t, binDir, "codex")
	t.Setenv("PATH", binDir)

	codexPath := filepath.Join(homeDir, ".codex", "skills", agentskill.SkillName)
	if err := os.MkdirAll(codexPath, 0o755); err != nil {
		t.Fatalf("create unmanaged skill dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(codexPath, "SKILL.md"), []byte("custom"), 0o644); err != nil {
		t.Fatalf("write unmanaged skill file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill", "install"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !strings.Contains(output, `exists but is not managed by sonacli`) {
		t.Fatalf("expected unmanaged skill error, got %q", output)
	}
}

func TestRunSkillUninstallRemovesInstalledSkillsByDefault(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	manager := agentskill.Manager{}
	results, err := manager.Install([]agentskill.Target{agentskill.TargetCodex, agentskill.TargetClaude})
	if err != nil {
		t.Fatalf("install managed skills: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill", "uninstall"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	want := "Removed sonacli skill for codex.\nPath: " + results[0].Path + "\n" +
		"Removed sonacli skill for claude.\nPath: " + results[1].Path + "\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	assertPathMissing(t, results[0].Path)
	assertPathMissing(t, results[1].Path)
}

func TestRunSkillUninstallReportsMissingSkills(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"skill", "uninstall"}, &stdout, &stderr)

	codexPath := filepath.Join(homeDir, ".codex", "skills", agentskill.SkillName)
	claudePath := filepath.Join(homeDir, ".claude", "skills", agentskill.SkillName)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", exitCode, stderr.String())
	}

	want := "sonacli skill is not installed for codex.\nPath: " + codexPath + "\n" +
		"sonacli skill is not installed for claude.\nPath: " + claudePath + "\n"
	if got := stdout.String(); got != want {
		t.Fatalf("unexpected stdout:\nwant %q\ngot  %q", want, got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func writeCLIExecutable(t *testing.T, dir, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable %q: %v", name, err)
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %q to exist, got %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %q to be missing, got %v", path, err)
	}
}
