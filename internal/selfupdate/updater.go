package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultRepo    = "mshddev/sonacli"
	defaultAPIBase = "https://api.github.com"
)

type Updater struct {
	Client         *http.Client
	Repo           string
	APIBase        string
	DownloadBase   string
	ExecutablePath string
	CurrentVersion string
	GOOS           string
	GOARCH         string
}

type Result struct {
	Version         string
	PreviousVersion string
	ExecutablePath  string
	Updated         bool
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func NewFromEnvironment(currentVersion string) (*Updater, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve current executable path: %w", err)
	}

	if resolvedPath, err := filepath.EvalSymlinks(executablePath); err == nil {
		executablePath = resolvedPath
	}

	repo := valueOrDefault(os.Getenv("SONACLI_INSTALL_REPO"), defaultRepo)
	downloadBase := os.Getenv("SONACLI_INSTALL_DOWNLOAD_BASE")
	if downloadBase == "" {
		downloadBase = fmt.Sprintf("https://github.com/%s/releases/download", repo)
	}

	return &Updater{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Repo:           repo,
		APIBase:        valueOrDefault(os.Getenv("SONACLI_INSTALL_API_BASE"), defaultAPIBase),
		DownloadBase:   downloadBase,
		ExecutablePath: executablePath,
		CurrentVersion: currentVersion,
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
	}, nil
}

func (u *Updater) Update(ctx context.Context, requestedVersion string) (Result, error) {
	targetVersion, err := u.resolveVersion(ctx, requestedVersion)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Version:         targetVersion,
		PreviousVersion: u.CurrentVersion,
		ExecutablePath:  u.ExecutablePath,
	}

	if targetVersion == u.CurrentVersion {
		return result, nil
	}

	assetBase, err := u.assetBaseName(targetVersion)
	if err != nil {
		return Result{}, err
	}

	tmpDir, err := os.MkdirTemp("", "sonacli-update-*")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archiveName := assetBase + ".tar.gz"
	archivePath := filepath.Join(tmpDir, archiveName)
	checksumsPath := filepath.Join(tmpDir, "checksums.txt")
	extractedBinaryPath := filepath.Join(tmpDir, "sonacli")

	if err := u.downloadToFile(ctx, u.releaseAssetURL(targetVersion, archiveName), archivePath); err != nil {
		return Result{}, err
	}

	if err := u.downloadToFile(ctx, u.releaseAssetURL(targetVersion, "checksums.txt"), checksumsPath); err != nil {
		return Result{}, err
	}

	if err := verifyChecksum(archiveName, archivePath, checksumsPath); err != nil {
		return Result{}, err
	}

	if err := extractBinaryFromArchive(archivePath, assetBase+"/sonacli", extractedBinaryPath); err != nil {
		return Result{}, err
	}

	if err := replaceExecutable(extractedBinaryPath, u.ExecutablePath); err != nil {
		return Result{}, err
	}

	result.Updated = true

	return result, nil
}

func (u *Updater) resolveVersion(ctx context.Context, requestedVersion string) (string, error) {
	if requestedVersion != "" {
		return requestedVersion, nil
	}

	latestURL := fmt.Sprintf("%s/repos/%s/releases/latest", strings.TrimRight(u.APIBase, "/"), u.Repo)
	if tag, err := u.fetchLatestReleaseTag(ctx, latestURL); err == nil && tag != "" {
		return tag, nil
	}

	releasesURL := fmt.Sprintf("%s/repos/%s/releases?per_page=1", strings.TrimRight(u.APIBase, "/"), u.Repo)
	if tag, err := u.fetchReleasesTag(ctx, releasesURL); err == nil && tag != "" {
		return tag, nil
	}

	return "", fmt.Errorf("could not determine a release tag from %s", u.Repo)
}

func (u *Updater) fetchLatestReleaseTag(ctx context.Context, url string) (string, error) {
	var release githubRelease
	if err := u.getJSON(ctx, url, &release); err != nil {
		return "", err
	}

	return strings.TrimSpace(release.TagName), nil
}

func (u *Updater) fetchReleasesTag(ctx context.Context, url string) (string, error) {
	var releases []githubRelease
	if err := u.getJSON(ctx, url, &releases); err != nil {
		return "", err
	}

	if len(releases) == 0 {
		return "", errors.New("no releases found")
	}

	return strings.TrimSpace(releases[0].TagName), nil
}

func (u *Updater) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request for %s: %w", url, err)
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %s", url, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}

	return nil
}

func (u *Updater) downloadToFile(ctx context.Context, url, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request for %s: %w", url, err)
	}

	resp, err := u.Client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %s", url, resp.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create %s: %w", destination, err)
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		file.Close()
		return fmt.Errorf("write %s: %w", destination, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", destination, err)
	}

	return nil
}

func (u *Updater) assetBaseName(version string) (string, error) {
	goos, err := releaseGOOS(u.GOOS)
	if err != nil {
		return "", err
	}

	goarch, err := releaseGOARCH(u.GOARCH)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("sonacli_%s_%s_%s", strings.TrimPrefix(version, "v"), goos, goarch), nil
}

func (u *Updater) releaseAssetURL(version, filename string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(u.DownloadBase, "/"), version, filename)
}

func releaseGOOS(goos string) (string, error) {
	switch goos {
	case "linux", "darwin":
		return goos, nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}
}

func releaseGOARCH(goarch string) (string, error) {
	switch goarch {
	case "amd64", "arm64":
		return goarch, nil
	default:
		return "", fmt.Errorf("unsupported architecture: %s", goarch)
	}
}

func verifyChecksum(assetName, archivePath, checksumsPath string) error {
	expected, err := checksumForAsset(assetName, checksumsPath)
	if err != nil {
		return err
	}

	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", archivePath, err)
	}

	actual := sha256.Sum256(archiveBytes)
	actualHex := hex.EncodeToString(actual[:])
	if expected != actualHex {
		return fmt.Errorf("checksum verification failed for %s", assetName)
	}

	return nil
}

func checksumForAsset(assetName, checksumsPath string) (string, error) {
	contents, err := os.ReadFile(checksumsPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", checksumsPath, err)
	}

	for _, line := range strings.Split(string(contents), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := strings.TrimPrefix(fields[1], "*")
		if name == assetName {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("could not find %s in checksums.txt", assetName)
}

func extractBinaryFromArchive(archivePath, entryName, destination string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open %s: %w", archivePath, err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip stream for %s: %w", archivePath, err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	cleanEntryName := path.Clean(entryName)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive %s: %w", archivePath, err)
		}

		if path.Clean(header.Name) != cleanEntryName {
			continue
		}

		mode := header.FileInfo().Mode().Perm()
		if mode&0o111 == 0 {
			mode = 0o755
		}

		out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			return fmt.Errorf("create %s: %w", destination, err)
		}

		if _, err := io.Copy(out, tarReader); err != nil {
			out.Close()
			return fmt.Errorf("extract %s: %w", destination, err)
		}

		if err := out.Close(); err != nil {
			return fmt.Errorf("close %s: %w", destination, err)
		}

		return nil
	}

	return fmt.Errorf("release archive did not contain %s", entryName)
}

func replaceExecutable(sourcePath, targetPath string) error {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", sourcePath, err)
	}

	targetDir := filepath.Dir(targetPath)
	replacement, err := os.CreateTemp(targetDir, ".sonacli-update-*")
	if err != nil {
		return fmt.Errorf("create replacement binary in %s: %w", targetDir, err)
	}

	replacementPath := replacement.Name()
	cleanupReplacement := true
	defer func() {
		if cleanupReplacement {
			_ = os.Remove(replacementPath)
		}
	}()

	source, err := os.Open(sourcePath)
	if err != nil {
		replacement.Close()
		return fmt.Errorf("open %s: %w", sourcePath, err)
	}

	if _, err := io.Copy(replacement, source); err != nil {
		source.Close()
		replacement.Close()
		return fmt.Errorf("stage replacement binary: %w", err)
	}

	if err := source.Close(); err != nil {
		replacement.Close()
		return fmt.Errorf("close %s: %w", sourcePath, err)
	}

	if err := replacement.Chmod(sourceInfo.Mode().Perm()); err != nil {
		replacement.Close()
		return fmt.Errorf("chmod %s: %w", replacementPath, err)
	}

	if err := replacement.Close(); err != nil {
		return fmt.Errorf("close %s: %w", replacementPath, err)
	}

	if err := os.Rename(replacementPath, targetPath); err != nil {
		return fmt.Errorf("replace %s: %w", targetPath, err)
	}

	cleanupReplacement = false

	return nil
}

func valueOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}
