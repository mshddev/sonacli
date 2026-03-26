#!/bin/sh

set -eu

script_dir="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "${script_dir}/.." && pwd)"

server_url="${SONACLI_TEST_SONARQUBE_URL:-}"
admin_login="${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN:-${SONACLI_SONARQUBE_ADMIN_LOGIN:-admin}}"
admin_password="${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD:-${SONACLI_SONARQUBE_ADMIN_PASSWORD:-SonacliAdmin1@}}"
project_key="${SONACLI_SAMPLE_PROJECT_KEY:-sonacli-sample-basic}"
project_name="${SONACLI_SAMPLE_PROJECT_NAME:-sonacli Sample Basic}"
sample_dir="${SONACLI_SAMPLE_PROJECT_DIR:-${repo_root}/tests/fixtures/sample-basic}"
scanner_version="${SONACLI_SAMPLE_SCANNER_VERSION:-8.0.1.6346}"
tools_dir="${SONACLI_SAMPLE_TOOLS_DIR:-${repo_root}/tests/bin/sample-project-tools}"
wait_timeout="${SONACLI_SAMPLE_PROJECT_TIMEOUT:-240}"

token_name=""
sample_token=""
scan_workspace=""
scanner_platform=""

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		printf '%s\n' "Missing required command: $1" >&2
		exit 1
	fi
}

usage() {
	cat <<'EOF'
Usage:
  ./scripts/seed-sample-project.sh

This helper seeds the repo-owned sample SonarQube project used for local
development and future live integration tests.

Optional environment overrides:
  SONACLI_TEST_SONARQUBE_URL           SonarQube base URL
  SONACLI_TEST_SONARQUBE_ADMIN_LOGIN   SonarQube admin login
  SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD SonarQube admin password
  SONACLI_SAMPLE_PROJECT_KEY           Sample project key
  SONACLI_SAMPLE_PROJECT_NAME          Sample project name
  SONACLI_SAMPLE_PROJECT_DIR           Sample project directory
  SONACLI_SAMPLE_SCANNER_VERSION       Pinned SonarScanner CLI version
  SONACLI_SAMPLE_TOOLS_DIR             Local cache directory for scanner downloads
  SONACLI_SAMPLE_PROJECT_TIMEOUT       Wait timeout in seconds for analysis completion
EOF
}

cleanup() {
	if [ -n "$token_name" ]; then
		curl -fsS -u "${admin_login}:${admin_password}" \
			-X POST "${server_url}/api/user_tokens/revoke" \
			--data-urlencode "name=${token_name}" >/dev/null 2>&1 || true
	fi

	if [ -n "$scan_workspace" ] && [ -d "$scan_workspace" ]; then
		rm -rf "$scan_workspace"
	fi
}

json_string_field() {
	printf '%s\n' "$1" | sed -n "s/.*\"$2\":\"\\([^\"]*\\)\".*/\\1/p"
}

normalize_server_url() {
	printf '%s' "${1%/}"
}

resolve_scanner_platform() {
	os_name="$(uname -s)"
	arch_name="$(uname -m)"

	case "${os_name}:${arch_name}" in
		Linux:x86_64)
			printf '%s\n' "linux-x64"
			;;
		Linux:aarch64 | Linux:arm64)
			printf '%s\n' "linux-aarch64"
			;;
		Darwin:x86_64)
			printf '%s\n' "macosx-x64"
			;;
		Darwin:arm64 | Darwin:aarch64)
			printf '%s\n' "macosx-aarch64"
			;;
		*)
			printf '%s\n' "Unsupported platform for SonarScanner CLI: ${os_name} ${arch_name}" >&2
			exit 1
			;;
	esac
}

ensure_scanner() {
	scanner_archive="sonar-scanner-cli-${scanner_version}-${scanner_platform}.zip"
	scanner_url="https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/${scanner_archive}"
	scanner_root="${tools_dir}/sonar-scanner-${scanner_version}-${scanner_platform}"
	scanner_bin="${scanner_root}/bin/sonar-scanner"
	download_dir="${tools_dir}/downloads"
	archive_path="${download_dir}/${scanner_archive}"

	if [ -x "$scanner_bin" ]; then
		printf '%s\n' "$scanner_bin"
		return 0
	fi

	mkdir -p "$download_dir"
	rm -rf "$scanner_root"

	printf '%s\n' "Downloading SonarScanner CLI ${scanner_version} (${scanner_platform}) ..." >&2
	curl -fsSL "$scanner_url" -o "$archive_path"

	printf '%s\n' "Installing SonarScanner CLI into ${scanner_root} ..." >&2
	unzip -q "$archive_path" -d "$tools_dir"

	printf '%s\n' "$scanner_bin"
}

wait_for_up() {
	printf '%s\n' "Waiting for SonarQube at ${server_url} ..."
	started_at="$(date +%s)"

	while :; do
		now="$(date +%s)"
		if [ $((now - started_at)) -ge "$wait_timeout" ]; then
			printf '%s\n' "Timed out waiting for SonarQube to report status UP." >&2
			exit 1
		fi

		response="$(curl -fsS "${server_url}/api/system/status" 2>/dev/null || true)"
		case "$response" in
			*'"status":"UP"'*)
				printf '%s\n' "SonarQube is ready at ${server_url}"
				return 0
				;;
		esac

		sleep 2
	done
}

reset_project() {
	printf '%s\n' "Resetting sample project ${project_key} ..."

	curl -fsS -u "${admin_login}:${admin_password}" \
		-X POST "${server_url}/api/projects/delete" \
		--data-urlencode "project=${project_key}" >/dev/null 2>&1 || true

	curl -fsS -u "${admin_login}:${admin_password}" \
		-X POST "${server_url}/api/projects/create" \
		--data-urlencode "project=${project_key}" \
		--data-urlencode "name=${project_name}" >/dev/null
}

generate_token() {
	token_name="sonacli-sample-seed-$(date +%s)-$$"

	response="$(
		curl -fsS -u "${admin_login}:${admin_password}" \
			-X POST "${server_url}/api/user_tokens/generate" \
			--data-urlencode "name=${token_name}"
	)"

	sample_token="$(json_string_field "$response" token)"
	if [ -z "$sample_token" ]; then
		printf '%s\n' "Failed to extract the generated SonarQube token from the API response." >&2
		exit 1
	fi
}

run_analysis() {
	scanner_bin="$1"
	sonar_user_home="${tools_dir}/sonar-user-home"
	scanner_workdir="${scan_workspace}/.scannerwork"
	report_file="${scanner_workdir}/report-task.txt"

	mkdir -p "$sonar_user_home" "$scanner_workdir"

	printf '%s\n' "Running SonarScanner against ${sample_dir} ..."
	(
		cd "$sample_dir"
		SONAR_HOST_URL="$server_url" \
		SONAR_TOKEN="$sample_token" \
		SONAR_USER_HOME="$sonar_user_home" \
		"$scanner_bin" \
			-Dsonar.projectKey="$project_key" \
			-Dsonar.projectName="$project_name" \
			-Dsonar.working.directory="$scanner_workdir"
	)

	if [ ! -f "$report_file" ]; then
		printf '%s\n' "SonarScanner did not produce ${report_file}." >&2
		exit 1
	fi

	ce_task_id="$(sed -n 's/^ceTaskId=//p' "$report_file")"
	ce_task_url="$(sed -n 's/^ceTaskUrl=//p' "$report_file")"
	dashboard_url="$(sed -n 's/^dashboardUrl=//p' "$report_file")"

	if [ -z "$ce_task_id" ] || [ -z "$ce_task_url" ] || [ -z "$dashboard_url" ]; then
		printf '%s\n' "SonarScanner metadata file is missing required task details." >&2
		exit 1
	fi

	printf '%s\n' "Waiting for analysis task ${ce_task_id} to finish ..."
	started_at="$(date +%s)"

	while :; do
		now="$(date +%s)"
		if [ $((now - started_at)) -ge "$wait_timeout" ]; then
			printf '%s\n' "Timed out waiting for sample project analysis to finish." >&2
			exit 1
		fi

		response="$(curl -fsS -u "${admin_login}:${admin_password}" "$ce_task_url")"

		case "$response" in
			*'"status":"SUCCESS"'*)
				printf '%s\n' "Sample project analysis completed successfully."
				printf '\n'
				printf '%s\n' "Project key: ${project_key}"
				printf '%s\n' "Dashboard URL: ${dashboard_url}"
				printf '%s\n' "Issues API: ${server_url}/api/issues/search?componentKeys=${project_key}"
				printf '%s\n' "Measures API: ${server_url}/api/measures/component?component=${project_key}&metricKeys=bugs,vulnerabilities,code_smells"
				return 0
				;;
			*'"status":"FAILED"'* | *'"status":"CANCELED"'*)
				printf '%s\n' "Sample project analysis did not complete successfully." >&2
				printf '%s\n' "$response" >&2
				exit 1
				;;
		esac

		sleep 2
	done
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
	usage
	exit 0
fi

if [ "$#" -ne 0 ]; then
	usage >&2
	exit 1
fi

trap cleanup EXIT HUP INT TERM

need_cmd curl
need_cmd unzip
need_cmd mktemp

if [ -z "$server_url" ]; then
	server_url="$("${script_dir}/local-sonarqube.sh" url)"
fi

server_url="$(normalize_server_url "$server_url")"
scanner_platform="$(resolve_scanner_platform)"

if [ ! -d "$sample_dir" ]; then
	printf '%s\n' "Sample project directory does not exist: ${sample_dir}" >&2
	exit 1
fi

"${script_dir}/local-sonarqube.sh" start
wait_for_up
reset_project
generate_token

scan_workspace="$(mktemp -d)"
run_analysis "$(ensure_scanner)"
