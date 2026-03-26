#!/bin/sh

set -eu

container_name="${SONACLI_SONARQUBE_CONTAINER_NAME:-sonacli-sonarqube}"
image="${SONACLI_SONARQUBE_IMAGE:-docker.io/library/sonarqube:25.10.0.114319-community}"
port="${SONACLI_SONARQUBE_PORT:-9000}"
start_timeout="${SONACLI_SONARQUBE_START_TIMEOUT:-240}"
base_url="http://127.0.0.1:${port}"
admin_login="${SONACLI_SONARQUBE_ADMIN_LOGIN:-admin}"
bootstrap_admin_password="${SONACLI_SONARQUBE_BOOTSTRAP_ADMIN_PASSWORD:-admin}"
local_admin_password="${SONACLI_SONARQUBE_ADMIN_PASSWORD:-SonacliAdmin1@}"

data_volume="${container_name}-data"
extensions_volume="${container_name}-extensions"
logs_volume="${container_name}-logs"

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		printf '%s\n' "Missing required command: $1" >&2
		exit 1
	fi
}

ensure_podman_ready() {
	if podman info >/dev/null 2>&1; then
		return 0
	fi

	if [ "$(uname -s)" = "Darwin" ]; then
		cat >&2 <<'EOF'
Podman is installed but not ready.
On macOS, the Podman machine must be created and started by the human operator before using this helper.
Make sure these succeed first:
  podman machine start podman-machine-default
  podman info
EOF
		exit 1
	fi

	cat >&2 <<'EOF'
Podman is installed but not ready.
Make sure `podman info` works for your user, then retry.
EOF
	exit 1
}

container_exists() {
	podman ps -a --format '{{.Names}}' | grep -Fx "$container_name" >/dev/null 2>&1
}

container_running() {
	podman ps --format '{{.Names}}' | grep -Fx "$container_name" >/dev/null 2>&1
}

ensure_volume() {
	if podman volume inspect "$1" >/dev/null 2>&1; then
		return 0
	fi

	podman volume create "$1" >/dev/null
}

wait_for_up() {
	printf '%s\n' "Waiting for SonarQube at ${base_url} ..."
	started_at="$(date +%s)"

	while :; do
		now="$(date +%s)"
		if [ $((now - started_at)) -ge "$start_timeout" ]; then
			printf '%s\n' "Timed out waiting for SonarQube to report status UP." >&2
			podman logs --tail 50 "$container_name" >&2 || true
			exit 1
		fi

		response="$(curl -fsS "${base_url}/api/system/status" 2>/dev/null || true)"
		case "$response" in
			*'"status":"UP"'*)
				printf '%s\n' "SonarQube is ready at ${base_url}"
				return 0
				;;
		esac

		sleep 2
	done
}

can_authenticate_admin() {
	curl -fsS \
		-u "${admin_login}:$1" \
		"${base_url}/api/users/current" >/dev/null 2>&1
}

change_local_admin_password() {
	curl -fsS \
		-u "${admin_login}:${bootstrap_admin_password}" \
		-X POST "${base_url}/api/users/change_password" \
		--data-urlencode "login=${admin_login}" \
		--data-urlencode "previousPassword=${bootstrap_admin_password}" \
		--data-urlencode "password=${local_admin_password}" >/dev/null
}

ensure_local_admin_password() {
	if can_authenticate_admin "$local_admin_password"; then
		printf '%s\n' "Local SonarQube testing credentials are ready: ${admin_login} / ${local_admin_password}"
		return 0
	fi

	if ! can_authenticate_admin "$bootstrap_admin_password"; then
		cat >&2 <<EOF
Local SonarQube is running, but the helper could not verify the documented testing password.
Expected local testing credentials:
  ${admin_login} / ${local_admin_password}

This usually means the existing SonarQube volumes were initialized earlier with a different admin password.
Change the SonarQube admin password to ${local_admin_password}, or remove the local container and named volumes and run:
  ./scripts/local-sonarqube.sh start
EOF
		exit 1
	fi

	printf '%s\n' "Setting local SonarQube admin password to testing value ..."
	change_local_admin_password

	if ! can_authenticate_admin "$local_admin_password"; then
		printf '%s\n' "Failed to verify the local SonarQube admin password after updating it." >&2
		exit 1
	fi

	printf '%s\n' "Local SonarQube testing credentials are ready: ${admin_login} / ${local_admin_password}"
}

start_container() {
	need_cmd podman
	need_cmd curl
	ensure_podman_ready

	if container_running; then
		printf '%s\n' "Container ${container_name} is already running."
		wait_for_up
		ensure_local_admin_password
		return 0
	fi

	if container_exists; then
		printf '%s\n' "Starting existing container ${container_name} ..."
		podman start "$container_name" >/dev/null
		wait_for_up
		ensure_local_admin_password
		return 0
	fi

	printf '%s\n' "Creating local SonarQube container ${container_name} ..."
	ensure_volume "$data_volume"
	ensure_volume "$extensions_volume"
	ensure_volume "$logs_volume"

	podman run -d \
		--name "$container_name" \
		-p "${port}:9000" \
		-v "${data_volume}:/opt/sonarqube/data" \
		-v "${extensions_volume}:/opt/sonarqube/extensions" \
		-v "${logs_volume}:/opt/sonarqube/logs" \
		"$image" >/dev/null

	wait_for_up
	ensure_local_admin_password
}

stop_container() {
	need_cmd podman
	ensure_podman_ready

	if ! container_exists; then
		printf '%s\n' "Container ${container_name} does not exist."
		return 0
	fi

	if ! container_running; then
		printf '%s\n' "Container ${container_name} is already stopped."
		return 0
	fi

	podman stop "$container_name" >/dev/null
	printf '%s\n' "Stopped ${container_name}."
}

remove_container() {
	need_cmd podman
	ensure_podman_ready

	if ! container_exists; then
		printf '%s\n' "Container ${container_name} does not exist."
		return 0
	fi

	podman rm -f "$container_name" >/dev/null
	printf '%s\n' "Removed ${container_name}. Named volumes were kept."
}

show_logs() {
	need_cmd podman
	ensure_podman_ready

	if ! container_exists; then
		printf '%s\n' "Container ${container_name} does not exist." >&2
		exit 1
	fi

	podman logs --tail 100 "$container_name"
}

show_status() {
	need_cmd podman
	ensure_podman_ready

	if ! container_exists; then
		printf '%s\n' "Container ${container_name} does not exist."
		exit 1
	fi

	podman ps -a \
		--filter "name=^${container_name}$" \
		--format 'table {{.Names}}\t{{.Status}}\t{{.Image}}\t{{.Ports}}'

	if container_running; then
		printf '\n%s\n' "API status:"
		curl -fsS "${base_url}/api/system/status"
		printf '\n'
	fi
}

show_url() {
	printf '%s\n' "$base_url"
}

usage() {
	cat <<'EOF'
Usage:
  ./scripts/local-sonarqube.sh <command>

Commands:
  start    Start the local SonarQube Community Build container and wait for it
  stop     Stop the local container
  rm       Remove the local container and keep named volumes
  logs     Show recent container logs
  status   Show container status and API status when running
  url      Print the local SonarQube base URL
EOF
}

command="${1:-}"

case "$command" in
	start)
		start_container
		;;
	stop)
		stop_container
		;;
	rm)
		remove_container
		;;
	logs)
		show_logs
		;;
	status)
		show_status
		;;
	url)
		show_url
		;;
	*)
		usage >&2
		exit 1
		;;
esac
