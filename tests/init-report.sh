#!/bin/sh

set -eu

usage() {
	cat <<'EOF' >&2
Usage:
  ./tests/init-report.sh <case-file> [output-file]

Examples:
  ./tests/init-report.sh 01-root.md
  ./tests/init-report.sh tests/01-root.md /tmp/01-root-testresults.md
EOF
}

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
	usage
	exit 1
fi

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)

case_input=$1
case_path=$case_input
if [ ! -f "$case_path" ]; then
	case_path="$script_dir/$(basename -- "$case_input")"
fi

if [ ! -f "$case_path" ]; then
	printf '%s\n' "Case file not found: $case_input" >&2
	exit 1
fi

case_file_name=$(basename -- "$case_path")
case_base=${case_file_name%.md}
if [ "$case_base" = "$case_file_name" ]; then
	printf '%s\n' "Case file must end with .md: $case_file_name" >&2
	exit 1
fi

output_path=${2:-"$script_dir/results/${case_base}-testresults.md"}
if [ -e "$output_path" ]; then
	printf '%s\n' "Refusing to overwrite existing file: $output_path" >&2
	exit 1
fi

report_title=$(sed -n 's/^# //p;q' "$case_path")
if [ -z "$report_title" ]; then
	printf '%s\n' "Could not read the top-level title from: $case_path" >&2
	exit 1
fi

tmp_cases=$(mktemp)
trap 'rm -f "$tmp_cases"' EXIT INT TERM HUP

awk '
/^### / {
	line = substr($0, 5)
	id = line
	sub(/ .*/, "", id)
	title = line
	sub(/^[^ ]+ /, "", title)
	printf "%s\t%s\n", id, title
}
' "$case_path" >"$tmp_cases"

if [ ! -s "$tmp_cases" ]; then
	printf '%s\n' "No case headings found in: $case_path" >&2
	exit 1
fi

selected_case_ids=$(
	awk '
	BEGIN { first = 1 }
	{
		if (!first) {
			printf ", "
		}
		printf "`%s`", $1
		first = 0
	}
	' "$tmp_cases"
)

mkdir -p "$(dirname -- "$output_path")"

{
	printf '# %s Test Results\n\n' "$report_title"
	printf '## Run Summary\n\n'
	printf -- '- Run date: %s\n' "$(date '+%Y-%m-%d')"
	printf -- '- Working directory: `%s`\n' "$repo_root"
	printf -- '- Case file executed: [%s](../%s)\n' "$case_file_name" "$case_file_name"
	printf -- '- Selected case IDs: %s\n' "$selected_case_ids"
	printf -- '- Overall result: `PENDING`\n\n'

	printf '## Shared Setup Used\n\n'
	printf -- '- Built the real CLI binary at `./tests/bin/sonacli`\n'
	printf -- '- Used one isolated temporary `HOME` for the full selected run\n'
	printf -- '- Ran all commands from the repository root\n'
	printf -- '- <additional shared setup or infrastructure used>\n'
	printf -- '- Cleaned up per-case temporary capture files after each case\n'
	printf -- '- Cleaned up the isolated `HOME` and removed `./tests/bin/sonacli` after the selected run\n\n'

	printf '## Case Results\n\n'
	printf '| Case ID | Result | Command | Exit Code | Stdout | Stderr | Filesystem |\n'
	printf '| --- | --- | --- | --- | --- | --- | --- |\n'
	while IFS="$(printf '\t')" read -r case_id case_title; do
		printf '| `%s` | `PENDING` | `<exact command>` | expected `<n>`, actual `<n>` | pending | pending | pending |\n' "$case_id"
	done <"$tmp_cases"
	printf '\n'

	printf '## Detailed Results\n\n'
	while IFS="$(printf '\t')" read -r case_id case_title; do
		printf '### %s %s\n\n' "$case_id" "$case_title"
		printf -- '- Command run: `<exact command>`\n'
		printf -- '- Status: `PENDING`\n'
		printf -- '- Exit code: expected `<n>`, actual `<n>`\n'
		printf -- '- Stdout: pending\n'
		printf -- '- Stderr: pending\n'
		printf -- '- Filesystem: pending\n'
		printf -- '- Unexpected output or exit-code mismatches: pending\n\n'

		printf 'Actual `stdout`:\n\n'
		printf '```text\n<pending>\n```\n\n'

		printf 'Actual `stderr`:\n\n'
		printf '```text\n<pending>\n```\n\n'

		printf 'Actual filesystem state:\n\n'
		printf '```text\n<pending>\n```\n\n'

		printf 'If the case failed, include mismatch details:\n\n'
		printf '```diff\n<pending>\n```\n\n'
	done <"$tmp_cases"

	printf '## Cleanup Verification\n\n'
	printf -- '- `./tests/bin/sonacli` removed after the run: pending\n'
	printf -- '- <case-specific cleanup verification, such as token revocation, if applicable>\n'
} >"$output_path"

printf '%s\n' "Initialized report skeleton: $output_path"
