package cli

import (
	"bytes"
	"testing"
)

func TestRunVersionCommandPrintsVersion(t *testing.T) {
	originalVersion := Version
	Version = "v1.2.3"
	t.Cleanup(func() {
		Version = originalVersion
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if got := stdout.String(); got != "sonacli v1.2.3\n" {
		t.Fatalf("expected version output, got %q", got)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}
