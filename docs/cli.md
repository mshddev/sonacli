# sonacli CLI Reference

`sonacli` reads SonarQube analysis reports and prints them as JSON.

Only SonarQube Community Build `v25.x+` is supported.

## Authentication

Before using any command, connect `sonacli` to your SonarQube instance.

### `sonacli auth setup`

Authenticate `sonacli` to your SonarQube instance by providing the server URL and a user token.

```sh
sonacli auth setup --server-url http://127.0.0.1:9000 --token <token>
# or with short flags:
sonacli auth setup -s http://127.0.0.1:9000 -t <token>
```

| Flag | Description |
|------|-------------|
| `--server-url`, `-s` | SonarQube server URL |
| `--token`, `-t` | SonarQube user token |

Credentials are saved to `~/.sonacli/config.yaml`.

### `sonacli auth status`

Check whether authentication is properly configured.

## Projects

### `sonacli project list`

List the projects on your SonarQube instance.

```sh
sonacli project list --pretty
sonacli project list --page 2 --page-size 10
```

| Flag | Description |
|------|-------------|
| `--pretty` | Human-readable output |
| `--page` | Page number (default 1) |
| `--page-size` | Results per page |

## Issues

### `sonacli issue list <project-key>`

List issues for a project. By default, issues are filtered by status `OPEN` and `CONFIRMED`.

```sh
sonacli issue list my-project --pretty
sonacli issue list my-project --status ACCEPTED,FIXED
sonacli issue list my-project --severity HIGH,BLOCKER --qualities SECURITY
sonacli issue list my-project --me
```

| Flag | Description |
|------|-------------|
| `--pretty` | Human-readable output |
| `--status`, `-s` | Filter by status: `OPEN`, `CONFIRMED`, `FALSE_POSITIVE`, `ACCEPTED`, `FIXED` (default `OPEN,CONFIRMED`) |
| `--severity`, `-e` | Filter by severity: `INFO`, `LOW`, `MEDIUM`, `HIGH`, `BLOCKER` |
| `--qualities`, `-q` | Filter by software quality: `MAINTAINABILITY`, `RELIABILITY`, `SECURITY` |
| `--assigned`, `-a` | Show only assigned (`true`) or unassigned (`false`) issues |
| `--assignees`, `-i` | Filter by assignee |
| `--me`, `-m` | Show only issues assigned to the current authenticated user (me) |
| `--page` | Page number (default 1) |
| `--page-size` | Results per page |

### `sonacli issue show <issue-key-or-url>`

Show a single issue. You can pass either a plain issue key or a SonarQube issue URL — `sonacli` extracts the issue key from the URL automatically.

```sh
sonacli issue show AX1234567890
sonacli issue show 'https://sonarqube.example.com/project/issues?id=my-project&issues=AX1234567890'
```

| Flag | Description |
|------|-------------|
| `--pretty` | Human-readable output |

## Skills

`sonacli` can install itself as a skill for AI coding agents.

### `sonacli skill install`

Install the `sonacli` skill for Claude Code, Codex, or both. By default it auto-detects which agents are available on your `PATH`.

```sh
sonacli skill install            # auto-detect
sonacli skill install --claude   # Claude Code only
sonacli skill install --codex    # Codex only
```

| Flag | Description |
|------|-------------|
| `--claude` | Install for Claude Code |
| `--codex` | Install for Codex |

### `sonacli skill uninstall`

Remove a previously installed skill. Only removes directories that were installed by `sonacli`.

```sh
sonacli skill uninstall
sonacli skill uninstall --claude
sonacli skill uninstall --codex
```

| Flag | Description |
|------|-------------|
| `--claude` | Remove from Claude Code |
| `--codex` | Remove from Codex |

## Repository

### `sonacli star`

Open the `sonacli` GitHub repository in your default browser so you can star it manually.

```sh
sonacli star
```

On macOS this uses `open`. On Linux this uses `xdg-open`.

## Updating

### `sonacli update`

Download the matching GitHub release archive for the current OS and CPU, verify `checksums.txt`, and replace the running `sonacli` executable in place.

```sh
sonacli update
sonacli update --version v0.1.0-rc.3
```

If the requested tag matches the current build version, `sonacli` reports that it is already installed and exits without replacing the binary.

| Flag | Description |
|------|-------------|
| `--version` | Install a specific release tag instead of the latest release |

## Versioning

### `sonacli version`

Print the build version. 
