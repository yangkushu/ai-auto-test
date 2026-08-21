#!/usr/bin/env sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
skill_dir="$project_root/.agents/skills/execute-test-cases"
version=$(tr -d '\r\n' < "$skill_dir/VERSION")

for target in \
    windows-amd64.exe \
    linux-amd64
do
    file="$skill_dir/bin/ai-auto-test-store-$target"
    if [ ! -s "$file" ]; then
        echo "missing or empty distribution file: $file" >&2
        exit 1
    fi
done

actual=$($skill_dir/bin/ai-auto-test-store-linux-amd64 version)
printf '%s\n' "$actual" | grep -F '"ok":true' >/dev/null
printf '%s\n' "$actual" | grep -F '"version":"'"$version"'"' >/dev/null
