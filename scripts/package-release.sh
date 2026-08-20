#!/usr/bin/env sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
skill_dir="$project_root/.agents/skills/execute-test-cases"
version=$(tr -d '\r\n' < "$skill_dir/VERSION")
dist_dir="$project_root/dist/v$version"
stage_dir=$(mktemp -d)

cleanup() {
    rm -rf "$stage_dir"
}
trap cleanup EXIT HUP INT TERM

command -v tar >/dev/null 2>&1 || {
    echo "tar is required to package the release" >&2
    exit 1
}
command -v sha256sum >/dev/null 2>&1 || {
    echo "sha256sum is required to package the release" >&2
    exit 1
}

mkdir -p "$dist_dir" "$stage_dir/execute-test-cases"
cp -R "$skill_dir/." "$stage_dir/execute-test-cases/"

archive="$dist_dir/execute-test-cases-v$version-all-platforms.tar.gz"
tar -czf "$archive" -C "$stage_dir" execute-test-cases

for file in "$skill_dir"/bin/ai-auto-test-store-*
do
    cp "$file" "$dist_dir/"
done

(
    cd "$dist_dir"
    sha256sum execute-test-cases-v"$version"-all-platforms.tar.gz ai-auto-test-store-* > SHA256SUMS
)
