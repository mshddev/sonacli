package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallScriptInstallsSpecificVersion(t *testing.T) {
	t.Parallel()

	server := newInstallTestServer(t, installTestConfig{
		tag:            "v1.2.3",
		latestTag:      "v1.2.3",
		releasesTag:    "v1.2.3",
		latestStatus:   http.StatusOK,
		releasesStatus: http.StatusOK,
	})
	installDir := t.TempDir()

	output := runInstallScript(t, server, "--version", "v1.2.3", "--install-dir", installDir)
	verifyInstalledBinary(t, installDir, "v1.2.3")

	if !strings.Contains(output, "installed sonacli v1.2.3 to "+filepath.Join(installDir, "sonacli")) {
		t.Fatalf("install output %q did not report the installed binary path", output)
	}
}

func TestInstallScriptFallsBackToFirstReleaseWhenLatestEndpointFails(t *testing.T) {
	t.Parallel()

	server := newInstallTestServer(t, installTestConfig{
		tag:            "v2.0.0-rc.1",
		releasesTag:    "v2.0.0-rc.1",
		latestStatus:   http.StatusNotFound,
		releasesStatus: http.StatusOK,
	})
	installDir := t.TempDir()

	runInstallScript(t, server, "--install-dir", installDir)
	verifyInstalledBinary(t, installDir, "v2.0.0-rc.1")
}

type installTestConfig struct {
	tag            string
	latestTag      string
	releasesTag    string
	latestStatus   int
	releasesStatus int
}

type installTestServer struct {
	apiBase      string
	downloadBase string
}

func newInstallTestServer(t *testing.T, cfg installTestConfig) installTestServer {
	t.Helper()

	goos, goarch := installReleaseTarget(t)
	assetBase := fmt.Sprintf("sonacli_%s_%s_%s", strings.TrimPrefix(cfg.tag, "v"), goos, goarch)
	archiveName := assetBase + ".tar.gz"
	binaryContents := "#!/bin/sh\nprintf '%s\\n' \"fake sonacli " + cfg.tag + "\"\n"
	archiveBytes := makeInstallArchive(t, assetBase, binaryContents)
	checksum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(checksum[:]), archiveName)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/mshddev/sonacli/releases/latest":
			if cfg.latestStatus != http.StatusOK {
				http.Error(w, http.StatusText(cfg.latestStatus), cfg.latestStatus)
				return
			}
			fmt.Fprintf(w, `{"tag_name":%q}`, cfg.latestTag)
		case r.URL.Path == "/repos/mshddev/sonacli/releases":
			if cfg.releasesStatus != http.StatusOK {
				http.Error(w, http.StatusText(cfg.releasesStatus), cfg.releasesStatus)
				return
			}
			fmt.Fprintf(w, `[{"tag_name":%q}]`, cfg.releasesTag)
		case r.URL.Path == "/download/"+cfg.tag+"/"+archiveName:
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write(archiveBytes)
		case r.URL.Path == "/download/"+cfg.tag+"/checksums.txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = io.WriteString(w, checksums)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return installTestServer{
		apiBase:      server.URL,
		downloadBase: server.URL + "/download",
	}
}

func runInstallScript(t *testing.T, server installTestServer, args ...string) string {
	t.Helper()

	scriptArgs := append([]string{"./install.sh"}, args...)
	cmd := exec.Command("sh", scriptArgs...)
	cmd.Env = append(os.Environ(),
		"SONACLI_INSTALL_API_BASE="+server.apiBase,
		"SONACLI_INSTALL_DOWNLOAD_BASE="+server.downloadBase,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed: %v\n%s", err, output)
	}

	return string(output)
}

func verifyInstalledBinary(t *testing.T, installDir, version string) {
	t.Helper()

	binaryPath := filepath.Join(installDir, "sonacli")
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("installed binary missing: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("installed binary %s is not executable", binaryPath)
	}

	cmd := exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("installed binary failed: %v\n%s", err, output)
	}

	if string(output) != "fake sonacli "+version+"\n" {
		t.Fatalf("installed binary output = %q, want %q", output, "fake sonacli "+version+"\n")
	}
}

func makeInstallArchive(t *testing.T, assetBase, binaryContents string) []byte {
	t.Helper()

	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, assetBase+".tar.gz")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}

	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)

	writeTarFile(t, tarWriter, assetBase+"/", 0o755, nil, tar.TypeDir)
	writeTarFile(t, tarWriter, assetBase+"/sonacli", 0o755, []byte(binaryContents), tar.TypeReg)

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive file: %v", err)
	}

	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	return archiveBytes
}

func writeTarFile(t *testing.T, tw *tar.Writer, name string, mode int64, content []byte, typeflag byte) {
	t.Helper()

	header := &tar.Header{
		Name:     name,
		Mode:     mode,
		Size:     int64(len(content)),
		Typeflag: typeflag,
	}
	if typeflag == tar.TypeDir {
		header.Size = 0
	}

	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write tar header for %s: %v", name, err)
	}
	if len(content) == 0 {
		return
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write tar contents for %s: %v", name, err)
	}
}

func installReleaseTarget(t *testing.T) (string, string) {
	t.Helper()

	var goos string
	switch runtime.GOOS {
	case "linux", "darwin":
		goos = runtime.GOOS
	default:
		t.Skipf("unsupported test GOOS %q", runtime.GOOS)
	}

	var goarch string
	switch runtime.GOARCH {
	case "amd64", "arm64":
		goarch = runtime.GOARCH
	default:
		t.Skipf("unsupported test GOARCH %q", runtime.GOARCH)
	}

	return goos, goarch
}
