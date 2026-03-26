package agentskill

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	SkillName         = "sonacli"
	markerFileName    = ".sonacli-managed"
	markerContents    = "managed by sonacli\n"
	directoryMode     = 0o755
	installedFileMode = 0o644
)

var ErrNoSupportedTargetsDetected = errors.New("no supported agent CLI detected on PATH: codex, claude; pass --codex or --claude to install explicitly")

type Target string

const (
	TargetCodex  Target = "codex"
	TargetClaude Target = "claude"
)

type Manager struct{}

type InstallResult struct {
	Target Target
	Path   string
}

type UninstallResult struct {
	Target  Target
	Path    string
	Removed bool
}

type targetState int

const (
	targetStateMissing targetState = iota
	targetStateManaged
	targetStateUnmanaged
)

type targetInspection struct {
	Target Target
	Path   string
	State  targetState
}

func SupportedTargets() []Target {
	return []Target{TargetCodex, TargetClaude}
}

func (Manager) ResolveInstallTargets(includeCodex, includeClaude bool) ([]Target, error) {
	targets := selectedTargets(includeCodex, includeClaude)
	if len(targets) > 0 {
		return targets, nil
	}

	detectedTargets := detectInstalledTargets()
	if len(detectedTargets) == 0 {
		return nil, ErrNoSupportedTargetsDetected
	}

	return detectedTargets, nil
}

func (Manager) ResolveUninstallTargets(includeCodex, includeClaude bool) []Target {
	targets := selectedTargets(includeCodex, includeClaude)
	if len(targets) > 0 {
		return targets
	}

	return SupportedTargets()
}

func (Manager) SkillDir(target Target) (string, error) {
	switch target {
	case TargetCodex:
		baseDir := strings.TrimSpace(os.Getenv("CODEX_HOME"))
		if baseDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve home directory: %w", err)
			}

			baseDir = filepath.Join(homeDir, ".codex")
		}

		return filepath.Join(baseDir, "skills", SkillName), nil
	case TargetClaude:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}

		return filepath.Join(homeDir, ".claude", "skills", SkillName), nil
	default:
		return "", fmt.Errorf("unsupported target: %s", target)
	}
}

func (m Manager) Install(targets []Target) ([]InstallResult, error) {
	inspections, err := m.inspectTargets(targets)
	if err != nil {
		return nil, err
	}

	if err := ensureManagedOrMissing(inspections); err != nil {
		return nil, err
	}

	results := make([]InstallResult, 0, len(inspections))
	for _, inspection := range inspections {
		if inspection.State == targetStateManaged {
			if err := os.RemoveAll(inspection.Path); err != nil {
				return nil, fmt.Errorf("remove existing managed %s skill: %w", inspection.Target, err)
			}
		}

		if err := writeManagedSkill(inspection.Target, inspection.Path); err != nil {
			return nil, fmt.Errorf("install %s skill: %w", inspection.Target, err)
		}

		results = append(results, InstallResult{
			Target: inspection.Target,
			Path:   inspection.Path,
		})
	}

	return results, nil
}

func (m Manager) Uninstall(targets []Target) ([]UninstallResult, error) {
	inspections, err := m.inspectTargets(targets)
	if err != nil {
		return nil, err
	}

	if err := ensureManagedOrMissing(inspections); err != nil {
		return nil, err
	}

	results := make([]UninstallResult, 0, len(inspections))
	for _, inspection := range inspections {
		removed := inspection.State == targetStateManaged
		if removed {
			if err := os.RemoveAll(inspection.Path); err != nil {
				return nil, fmt.Errorf("remove %s skill: %w", inspection.Target, err)
			}
		}

		results = append(results, UninstallResult{
			Target:  inspection.Target,
			Path:    inspection.Path,
			Removed: removed,
		})
	}

	return results, nil
}

func (m Manager) inspectTargets(targets []Target) ([]targetInspection, error) {
	inspections := make([]targetInspection, 0, len(targets))
	for _, target := range targets {
		inspection, err := m.inspectTarget(target)
		if err != nil {
			return nil, err
		}

		inspections = append(inspections, inspection)
	}

	return inspections, nil
}

func (m Manager) inspectTarget(target Target) (targetInspection, error) {
	path, err := m.SkillDir(target)
	if err != nil {
		return targetInspection{}, err
	}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return targetInspection{Target: target, Path: path, State: targetStateMissing}, nil
		}

		return targetInspection{}, fmt.Errorf("stat %s skill path: %w", target, err)
	}

	if !info.IsDir() {
		return targetInspection{Target: target, Path: path, State: targetStateUnmanaged}, nil
	}

	markerPath := filepath.Join(path, markerFileName)
	if _, err := os.Stat(markerPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return targetInspection{Target: target, Path: path, State: targetStateUnmanaged}, nil
		}

		return targetInspection{}, fmt.Errorf("stat %s skill marker: %w", target, err)
	}

	return targetInspection{Target: target, Path: path, State: targetStateManaged}, nil
}

func selectedTargets(includeCodex, includeClaude bool) []Target {
	targets := make([]Target, 0, 2)
	if includeCodex {
		targets = append(targets, TargetCodex)
	}
	if includeClaude {
		targets = append(targets, TargetClaude)
	}

	return targets
}

func detectInstalledTargets() []Target {
	targets := make([]Target, 0, 2)
	for _, target := range SupportedTargets() {
		if _, err := exec.LookPath(string(target)); err == nil {
			targets = append(targets, target)
		}
	}

	return targets
}

func ensureManagedOrMissing(inspections []targetInspection) error {
	for _, inspection := range inspections {
		if inspection.State == targetStateUnmanaged {
			return unmanagedSkillError{
				Target: inspection.Target,
				Path:   inspection.Path,
			}
		}
	}

	return nil
}

func writeManagedSkill(target Target, root string) error {
	for _, asset := range embeddedSkillAssets {
		if len(asset.targets) > 0 && !containsTarget(asset.targets, target) {
			continue
		}

		path := filepath.Join(root, asset.path)
		if err := os.MkdirAll(filepath.Dir(path), directoryMode); err != nil {
			return fmt.Errorf("create skill directory: %w", err)
		}

		data, err := embeddedAssets.ReadFile(asset.source)
		if err != nil {
			return fmt.Errorf("read embedded asset %q: %w", asset.source, err)
		}

		if err := os.WriteFile(path, data, installedFileMode); err != nil {
			return fmt.Errorf("write %q: %w", asset.path, err)
		}
	}

	markerPath := filepath.Join(root, markerFileName)
	if err := os.WriteFile(markerPath, []byte(markerContents), installedFileMode); err != nil {
		return fmt.Errorf("write skill marker: %w", err)
	}

	return nil
}

func containsTarget(targets []Target, t Target) bool {
	for _, candidate := range targets {
		if candidate == t {
			return true
		}
	}

	return false
}

type unmanagedSkillError struct {
	Target Target
	Path   string
}

func (e unmanagedSkillError) Error() string {
	return fmt.Sprintf("%s skill path %q exists but is not managed by sonacli", e.Target, e.Path)
}
