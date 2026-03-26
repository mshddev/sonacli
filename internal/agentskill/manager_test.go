package agentskill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveInstallTargetsDetectsInstalledCommands(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "codex")
	t.Setenv("PATH", binDir)

	manager := Manager{}

	targets, err := manager.ResolveInstallTargets(false, false)
	if err != nil {
		t.Fatalf("resolve install targets: %v", err)
	}

	if len(targets) != 1 || targets[0] != TargetCodex {
		t.Fatalf("unexpected targets: %#v", targets)
	}
}

func TestResolveInstallTargetsReturnsErrorWhenNothingDetected(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")
	t.Setenv("PATH", t.TempDir())

	manager := Manager{}

	_, err := manager.ResolveInstallTargets(false, false)
	if err == nil {
		t.Fatal("expected detection error")
	}

	if err != ErrNoSupportedTargetsDetected {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveInstallTargetsUsesExplicitSelectionWithoutDetection(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")
	t.Setenv("PATH", t.TempDir())

	manager := Manager{}

	targets, err := manager.ResolveInstallTargets(true, false)
	if err != nil {
		t.Fatalf("resolve install targets: %v", err)
	}

	if len(targets) != 1 || targets[0] != TargetCodex {
		t.Fatalf("unexpected targets: %#v", targets)
	}
}

func TestResolveUninstallTargetsDefaultsToBoth(t *testing.T) {
	manager := Manager{}

	targets := manager.ResolveUninstallTargets(false, false)
	if len(targets) != 2 || targets[0] != TargetCodex || targets[1] != TargetClaude {
		t.Fatalf("unexpected targets: %#v", targets)
	}
}

func TestSkillDirUsesCODEXHOMEWhenSet(t *testing.T) {
	homeDir := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexHome)

	manager := Manager{}

	path, err := manager.SkillDir(TargetCodex)
	if err != nil {
		t.Fatalf("resolve Codex skill dir: %v", err)
	}

	want := filepath.Join(codexHome, "skills", SkillName)
	if path != want {
		t.Fatalf("unexpected path:\nwant %q\ngot  %q", want, path)
	}
}

func TestInstallWritesManagedSkillFiles(t *testing.T) {
	homeDir := t.TempDir()
	codexHome := filepath.Join(t.TempDir(), "codex-home")
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", codexHome)

	manager := Manager{}

	results, err := manager.Install([]Target{TargetCodex, TargetClaude})
	if err != nil {
		t.Fatalf("install skills: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("unexpected result count: %#v", results)
	}

	for _, result := range results {
		skillFile := filepath.Join(result.Path, "SKILL.md")
		markerFile := filepath.Join(result.Path, markerFileName)

		assertFileContains(t, skillFile, "name: sonacli")
		assertFileContains(t, markerFile, "managed by sonacli")

		openAIFile := filepath.Join(result.Path, "agents", "openai.yaml")
		if result.Target == TargetCodex {
			assertFileContains(t, openAIFile, `display_name: "Sonacli"`)
		} else {
			if _, err := os.Stat(openAIFile); !os.IsNotExist(err) {
				t.Fatalf("expected %q to not exist for %s target, got err=%v", openAIFile, result.Target, err)
			}
		}
	}
}

func TestInstallRejectsUnmanagedExistingSkillDirectory(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	manager := Manager{}

	path, err := manager.SkillDir(TargetClaude)
	if err != nil {
		t.Fatalf("resolve skill dir: %v", err)
	}

	if err := os.MkdirAll(path, directoryMode); err != nil {
		t.Fatalf("create unmanaged skill dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("custom"), installedFileMode); err != nil {
		t.Fatalf("write unmanaged skill file: %v", err)
	}

	_, err = manager.Install([]Target{TargetClaude})
	if err == nil {
		t.Fatal("expected unmanaged skill error")
	}

	if !strings.Contains(err.Error(), "is not managed by sonacli") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUninstallRemovesManagedSkillDirectory(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	manager := Manager{}

	results, err := manager.Install([]Target{TargetClaude})
	if err != nil {
		t.Fatalf("install skill: %v", err)
	}

	removed, err := manager.Uninstall([]Target{TargetClaude})
	if err != nil {
		t.Fatalf("uninstall skill: %v", err)
	}

	if len(removed) != 1 || !removed[0].Removed {
		t.Fatalf("unexpected uninstall result: %#v", removed)
	}

	if _, err := os.Stat(results[0].Path); !os.IsNotExist(err) {
		t.Fatalf("expected skill dir to be removed, got err=%v", err)
	}
}

func TestUninstallRejectsUnmanagedExistingSkillDirectory(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("CODEX_HOME", "")

	manager := Manager{}

	path, err := manager.SkillDir(TargetClaude)
	if err != nil {
		t.Fatalf("resolve skill dir: %v", err)
	}

	if err := os.MkdirAll(path, directoryMode); err != nil {
		t.Fatalf("create unmanaged skill dir: %v", err)
	}

	_, err = manager.Uninstall([]Target{TargetClaude})
	if err == nil {
		t.Fatal("expected unmanaged skill error")
	}

	if !strings.Contains(err.Error(), "is not managed by sonacli") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeExecutable(t *testing.T, dir, name string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable %q: %v", name, err)
	}
}

func assertFileContains(t *testing.T, path, substring string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}

	if !strings.Contains(string(data), substring) {
		t.Fatalf("expected %q to contain %q, got %q", path, substring, string(data))
	}
}
