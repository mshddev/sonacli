package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	appDirName  = ".sonacli"
	fileName    = "config.yaml"
	dirMode     = 0o700
	fileMode    = 0o600
	tempPattern = ".config.yaml.tmp-*"
)

var ErrAuthSetupNotFound = errors.New("auth setup not found")

type AuthSetup struct {
	ServerURL string
	Token     string
}

type fileConfig struct {
	ServerURL string
	Token     string
}

func Path() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, appDirName, fileName), nil
}

func SaveAuthSetup(serverURL, token string) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	cfg := fileConfig{
		ServerURL: serverURL,
		Token:     token,
	}

	if err := save(path, cfg); err != nil {
		return "", err
	}

	return path, nil
}

func LoadAuthSetup() (AuthSetup, error) {
	path, err := Path()
	if err != nil {
		return AuthSetup{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AuthSetup{}, ErrAuthSetupNotFound
		}

		return AuthSetup{}, fmt.Errorf("read config file: %w", err)
	}

	setup, err := parseAuthSetup(data)
	if err != nil {
		return AuthSetup{}, fmt.Errorf("parse config file: %w", err)
	}

	return setup, nil
}

func save(path string, cfg fileConfig) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.Chmod(dir, dirMode); err != nil {
		return fmt.Errorf("set config directory permissions: %w", err)
	}

	tempFile, err := os.CreateTemp(dir, tempPattern)
	if err != nil {
		return fmt.Errorf("create temporary config file: %w", err)
	}

	tempPath := tempFile.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.WriteString(render(cfg)); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("write config file: %w", err)
	}

	if err := tempFile.Chmod(fileMode); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("set temporary config file permissions: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temporary config file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("move config file into place: %w", err)
	}

	if err := os.Chmod(path, fileMode); err != nil {
		return fmt.Errorf("set config file permissions: %w", err)
	}

	return nil
}

func render(cfg fileConfig) string {
	return fmt.Sprintf(
		"server_url: %s\ntoken: %s\n",
		strconv.Quote(cfg.ServerURL),
		strconv.Quote(cfg.Token),
	)
}

func parseAuthSetup(data []byte) (AuthSetup, error) {
	var setup AuthSetup
	var haveServerURL bool
	var haveToken bool

	for lineNumber, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSuffix(rawLine, "\r")
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		key, rawValue, ok := strings.Cut(line, ":")
		if !ok {
			return AuthSetup{}, fmt.Errorf("line %d: missing ':' separator", lineNumber+1)
		}

		key = strings.TrimSpace(key)
		rawValue = strings.TrimSpace(rawValue)

		switch key {
		case "server_url":
			if haveServerURL {
				return AuthSetup{}, fmt.Errorf("line %d: duplicate %q key", lineNumber+1, key)
			}

			value, err := strconv.Unquote(rawValue)
			if err != nil {
				return AuthSetup{}, fmt.Errorf("line %d: decode %q value: %w", lineNumber+1, key, err)
			}
			if strings.TrimSpace(value) == "" {
				return AuthSetup{}, fmt.Errorf("line %d: %q must not be empty", lineNumber+1, key)
			}

			setup.ServerURL = value
			haveServerURL = true
		case "token":
			if haveToken {
				return AuthSetup{}, fmt.Errorf("line %d: duplicate %q key", lineNumber+1, key)
			}

			value, err := strconv.Unquote(rawValue)
			if err != nil {
				return AuthSetup{}, fmt.Errorf("line %d: decode %q value: %w", lineNumber+1, key, err)
			}
			if strings.TrimSpace(value) == "" {
				return AuthSetup{}, fmt.Errorf("line %d: %q must not be empty", lineNumber+1, key)
			}

			setup.Token = value
			haveToken = true
		default:
			continue
		}
	}

	if !haveServerURL {
		return AuthSetup{}, errors.New(`missing "server_url"`)
	}

	if !haveToken {
		return AuthSetup{}, errors.New(`missing "token"`)
	}

	return setup, nil
}
