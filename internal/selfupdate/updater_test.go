package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
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
	"sync/atomic"
	"testing"
)

func TestUpdaterInstallsLatestRelease(t *testing.T) {
	t.Parallel()

	server := newTestServer(t, updateServerConfig{
		tag:            "v1.2.3",
		latestTag:      "v1.2.3",
		releasesTag:    "v1.2.3",
		latestStatus:   http.StatusOK,
		releasesStatus: http.StatusOK,
	})

	executablePath := createExecutable(t, "sonacli current version")
	updater := &Updater{
		Client:         server.client(),
		Repo:           "mshddev/sonacli",
		APIBase:        server.apiBase,
		DownloadBase:   server.downloadBase,
		ExecutablePath: executablePath,
		CurrentVersion: "v1.0.0",
		GOOS:           server.goos,
		GOARCH:         server.goarch,
	}

	result, err := updater.Update(context.Background(), "")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if !result.Updated {
		t.Fatalf("expected updater to replace the executable")
	}

	if result.Version != "v1.2.3" {
		t.Fatalf("updated version = %q, want %q", result.Version, "v1.2.3")
	}

	verifyExecutableOutput(t, executablePath, "fake sonacli v1.2.3\n")
}

func TestUpdaterFallsBackToFirstReleaseWhenLatestEndpointFails(t *testing.T) {
	t.Parallel()

	server := newTestServer(t, updateServerConfig{
		tag:            "v2.0.0-rc.1",
		releasesTag:    "v2.0.0-rc.1",
		latestStatus:   http.StatusNotFound,
		releasesStatus: http.StatusOK,
	})

	executablePath := createExecutable(t, "sonacli current version")
	updater := &Updater{
		Client:         server.client(),
		Repo:           "mshddev/sonacli",
		APIBase:        server.apiBase,
		DownloadBase:   server.downloadBase,
		ExecutablePath: executablePath,
		CurrentVersion: "v1.0.0",
		GOOS:           server.goos,
		GOARCH:         server.goarch,
	}

	result, err := updater.Update(context.Background(), "")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if result.Version != "v2.0.0-rc.1" {
		t.Fatalf("updated version = %q, want %q", result.Version, "v2.0.0-rc.1")
	}

	verifyExecutableOutput(t, executablePath, "fake sonacli v2.0.0-rc.1\n")
}

func TestUpdaterSkipsReplacementWhenVersionMatches(t *testing.T) {
	t.Parallel()

	server := newTestServer(t, updateServerConfig{
		tag:            "v1.2.3",
		latestTag:      "v1.2.3",
		latestStatus:   http.StatusOK,
		releasesStatus: http.StatusInternalServerError,
	})

	executablePath := createExecutable(t, "sonacli current version")
	updater := &Updater{
		Client:         server.client(),
		Repo:           "mshddev/sonacli",
		APIBase:        server.apiBase,
		DownloadBase:   server.downloadBase,
		ExecutablePath: executablePath,
		CurrentVersion: "v1.2.3",
		GOOS:           server.goos,
		GOARCH:         server.goarch,
	}

	result, err := updater.Update(context.Background(), "")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if result.Updated {
		t.Fatalf("expected updater to skip replacing the executable")
	}

	verifyExecutableOutput(t, executablePath, "sonacli current version\n")
	if server.downloadCount() != 1 {
		t.Fatalf("download request count = %d, want %d", server.downloadCount(), 1)
	}
}

type updateServerConfig struct {
	tag            string
	latestTag      string
	releasesTag    string
	latestStatus   int
	releasesStatus int
}

type updateTestServer struct {
	apiBase        string
	downloadBase   string
	goos           string
	goarch         string
	server         *httptest.Server
	requestCounter *requestCounter
}

type requestCounter struct {
	count atomic.Int64
}

func (c *requestCounter) add() {
	c.count.Add(1)
}

func (c *requestCounter) value() int {
	return int(c.count.Load())
}

func newTestServer(t *testing.T, cfg updateServerConfig) updateTestServer {
	t.Helper()

	goos, goarch := releaseTargetForTests(t)
	assetBase := fmt.Sprintf("sonacli_%s_%s_%s", strings.TrimPrefix(cfg.tag, "v"), goos, goarch)
	archiveName := assetBase + ".tar.gz"
	archiveBytes := makeArchive(t, assetBase, "#!/bin/sh\nprintf '%s\\n' \"fake sonacli "+cfg.tag+"\"\n")
	checksum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(checksum[:]), archiveName)
	counter := &requestCounter{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.add()

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

	return updateTestServer{
		apiBase:        server.URL,
		downloadBase:   server.URL + "/download",
		goos:           goos,
		goarch:         goarch,
		server:         server,
		requestCounter: counter,
	}
}

func (s updateTestServer) client() *http.Client {
	return s.server.Client()
}

func (s updateTestServer) downloadCount() int {
	return s.requestCounter.value()
}

func createExecutable(t *testing.T, contents string) string {
	t.Helper()

	binaryPath := filepath.Join(t.TempDir(), "sonacli")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' %q\n", contents)
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake executable: %v", err)
	}

	return binaryPath
}

func verifyExecutableOutput(t *testing.T, binaryPath, want string) {
	t.Helper()

	output, err := runExecutable(binaryPath)
	if err != nil {
		t.Fatalf("run executable: %v", err)
	}

	if output != want {
		t.Fatalf("executable output = %q, want %q", output, want)
	}
}

func runExecutable(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func makeArchive(t *testing.T, assetBase, binaryContents string) []byte {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), assetBase+".tar.gz")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}

	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)

	writeTarEntry(t, tarWriter, assetBase+"/", 0o755, nil, tar.TypeDir)
	writeTarEntry(t, tarWriter, assetBase+"/sonacli", 0o755, []byte(binaryContents), tar.TypeReg)

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive file: %v", err)
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}

	return data
}

func writeTarEntry(t *testing.T, tw *tar.Writer, name string, mode int64, content []byte, typeflag byte) {
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

func releaseTargetForTests(t *testing.T) (string, string) {
	t.Helper()

	switch runtime.GOOS {
	case "linux", "darwin":
	default:
		t.Skipf("unsupported GOOS %q", runtime.GOOS)
	}

	switch runtime.GOARCH {
	case "amd64", "arm64":
	default:
		t.Skipf("unsupported GOARCH %q", runtime.GOARCH)
	}

	return runtime.GOOS, runtime.GOARCH
}
