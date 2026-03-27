package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestRunStarCommandOpensRepositoryURL(t *testing.T) {
	originalFactory := newStarOpener
	t.Cleanup(func() {
		newStarOpener = originalFactory
	})

	opener := &stubBrowserOpener{}
	newStarOpener = func() (browserOpener, error) {
		return opener, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"star"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if opener.url != repoURL {
		t.Fatalf("opened URL = %q, want %q", opener.url, repoURL)
	}

	if got := stdout.String(); got != "Opened "+repoURL+" in your browser.\n" {
		t.Fatalf("stdout = %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunStarCommandShowsHelpWhenOpenFails(t *testing.T) {
	originalFactory := newStarOpener
	t.Cleanup(func() {
		newStarOpener = originalFactory
	})

	newStarOpener = func() (browserOpener, error) {
		return &stubBrowserOpener{err: errors.New("boom")}, nil
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"star"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}

	output := stderr.String()
	if !bytes.Contains(stderr.Bytes(), []byte("Error: boom")) {
		t.Fatalf("stderr = %q", output)
	}

	if !bytes.Contains(stderr.Bytes(), []byte("sonacli star")) {
		t.Fatalf("stderr = %q", output)
	}
}

func TestNewSystemBrowserOpenerSelectsCommandByOS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		goos    string
		command string
	}{
		{goos: "darwin", command: "open"},
		{goos: "linux", command: "xdg-open"},
	}

	for _, tc := range testCases {
		t.Run(tc.goos, func(t *testing.T) {
			t.Parallel()

			opener, err := newSystemBrowserOpener(tc.goos)
			if err != nil {
				t.Fatalf("newSystemBrowserOpener(%q) returned error: %v", tc.goos, err)
			}

			if opener.command != tc.command {
				t.Fatalf("command = %q, want %q", opener.command, tc.command)
			}
		})
	}
}

func TestNewSystemBrowserOpenerRejectsUnsupportedOS(t *testing.T) {
	t.Parallel()

	opener, err := newSystemBrowserOpener("windows")
	if err == nil {
		t.Fatalf("expected error, got opener %+v", opener)
	}
}

type stubBrowserOpener struct {
	url string
	err error
}

func (s *stubBrowserOpener) Open(_ context.Context, url string) error {
	s.url = url
	return s.err
}
