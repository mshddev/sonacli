#!/bin/sh

set -eu

usage() {
	cat <<'EOF'
Usage:
  ./scripts/changelog-release-notes.sh <version> [changelog-path]

Extract the body of a versioned CHANGELOG section for use as release notes.
The version must match the changelog heading exactly, for example:
  ## [v1.2.3] - 2026-03-27
EOF
}

version="${1:-}"
changelog_path="${2:-CHANGELOG.md}"

if [ -z "$version" ]; then
	usage >&2
	exit 1
fi

if [ ! -f "$changelog_path" ]; then
	printf '%s\n' "Missing changelog file: $changelog_path" >&2
	exit 1
fi

awk -v version="$version" '
BEGIN {
	heading = "## [" version "]"
	in_section = 0
	found = 0
	has_content = 0
}
index($0, heading) == 1 {
	suffix = substr($0, length(heading) + 1)
	if (suffix == "" || suffix ~ /^ - /) {
		in_section = 1
		found = 1
		next
	}
}
in_section && /^## \[/ {
	exit
}
in_section && /^\[[^]]+\]: / {
	exit
}
in_section {
	if (!has_content && $0 == "") {
		next
	}
	has_content = 1
	print
}
END {
	if (!found) {
		printf "Release heading not found for %s in %s\n", version, FILENAME > "/dev/stderr"
		exit 2
	}
	if (!has_content) {
		printf "Release section for %s is empty in %s\n", version, FILENAME > "/dev/stderr"
		exit 3
	}
}
' "$changelog_path"
