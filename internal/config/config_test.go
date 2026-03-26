package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveWritesExpectedConfigFile(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	path := filepath.Join(rootDir, ".sonacli", "config.yaml")

	cfg := fileConfig{
		ServerURL: "http://127.0.0.1:9000",
		Token:     "test-token",
	}

	if err := save(path, cfg); err != nil {
		t.Fatalf("save returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}

	want := "server_url: \"http://127.0.0.1:9000\"\ntoken: \"test-token\"\n"
	if got := string(data); got != want {
		t.Fatalf("unexpected config contents:\nwant %q\ngot  %q", want, got)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}

	if got := info.Mode().Perm(); got != fileMode {
		t.Fatalf("unexpected config file mode: want %o, got %o", fileMode, got)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat config directory: %v", err)
	}

	if got := dirInfo.Mode().Perm(); got != dirMode {
		t.Fatalf("unexpected config directory mode: want %o, got %o", dirMode, got)
	}
}

func TestPathUsesDotSonacliDirectoryInHome(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	path, err := Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	want := filepath.Join(homeDir, ".sonacli", "config.yaml")
	if path != want {
		t.Fatalf("unexpected config path:\nwant %q\ngot  %q", want, path)
	}
}

func TestLoadAuthSetupReadsSavedConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	if _, err := SaveAuthSetup("http://127.0.0.1:9000", "test-token"); err != nil {
		t.Fatalf("save auth setup: %v", err)
	}

	setup, err := LoadAuthSetup()
	if err != nil {
		t.Fatalf("load auth setup: %v", err)
	}

	if setup.ServerURL != "http://127.0.0.1:9000" {
		t.Fatalf("unexpected server URL: %q", setup.ServerURL)
	}

	if setup.Token != "test-token" {
		t.Fatalf("unexpected token: %q", setup.Token)
	}
}

func TestLoadAuthSetupReturnsNotFoundWhenConfigIsMissing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	_, err := LoadAuthSetup()
	if !errors.Is(err, ErrAuthSetupNotFound) {
		t.Fatalf("expected ErrAuthSetupNotFound, got %v", err)
	}
}

func TestLoadAuthSetupRejectsMalformedConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	path, err := Path()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	if err := os.WriteFile(path, []byte("server_url: http://127.0.0.1:9000\ntoken: \"test-token\"\n"), fileMode); err != nil {
		t.Fatalf("write malformed config: %v", err)
	}

	_, err = LoadAuthSetup()
	if err == nil {
		t.Fatal("expected malformed config error")
	}

	if !strings.Contains(err.Error(), `decode "server_url" value`) {
		t.Fatalf("expected decode error, got %v", err)
	}
}
