package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/mshddev/sonacli/internal/selfupdate"
)

func TestRunUpdateCommandPrintsSuccessMessage(t *testing.T) {
	originalFactory := newUpdateRunner
	t.Cleanup(func() {
		newUpdateRunner = originalFactory
	})

	runner := &stubUpdateRunner{
		result: selfupdate.Result{
			Version:         "v1.2.3",
			PreviousVersion: "v1.0.0",
			ExecutablePath:  "/tmp/sonacli",
			Updated:         true,
		},
	}

	newUpdateRunner = func() (updateRunner, error) {
		return runner, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"update", "--version", "v1.2.3"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if runner.requestedVersion != "v1.2.3" {
		t.Fatalf("requested version = %q, want %q", runner.requestedVersion, "v1.2.3")
	}

	if got := stdout.String(); got != "Updated sonacli from v1.0.0 to v1.2.3.\nPath: /tmp/sonacli\n" {
		t.Fatalf("stdout = %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunUpdateCommandPrintsAlreadyInstalledMessage(t *testing.T) {
	originalFactory := newUpdateRunner
	t.Cleanup(func() {
		newUpdateRunner = originalFactory
	})

	newUpdateRunner = func() (updateRunner, error) {
		return &stubUpdateRunner{
			result: selfupdate.Result{
				Version:        "v1.2.3",
				ExecutablePath: "/tmp/sonacli",
			},
		}, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"update"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if got := stdout.String(); got != "sonacli v1.2.3 is already installed.\nPath: /tmp/sonacli\n" {
		t.Fatalf("stdout = %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunUpdateCommandShowsHelpWhenUpdaterFails(t *testing.T) {
	originalFactory := newUpdateRunner
	t.Cleanup(func() {
		newUpdateRunner = originalFactory
	})

	newUpdateRunner = func() (updateRunner, error) {
		return &stubUpdateRunner{
			err: errors.New("boom"),
		}, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"update"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if output == "" {
		t.Fatalf("expected stderr output")
	}

	if !bytes.Contains(stderr.Bytes(), []byte("Error: boom")) {
		t.Fatalf("stderr = %q", output)
	}

	if !bytes.Contains(stderr.Bytes(), []byte("sonacli update [flags]")) {
		t.Fatalf("stderr = %q", output)
	}
}

type stubUpdateRunner struct {
	requestedVersion string
	result           selfupdate.Result
	err              error
}

func (s *stubUpdateRunner) Update(_ context.Context, requestedVersion string) (selfupdate.Result, error) {
	s.requestedVersion = requestedVersion
	if s.err != nil {
		return selfupdate.Result{}, s.err
	}

	return s.result, nil
}
