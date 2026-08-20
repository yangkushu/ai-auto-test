#!/usr/bin/env sh
set -eu

project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=$(tr -d '\r\n' < "$project_root/.agents/skills/execute-test-cases/VERSION")
required_go_version=$(tr -d '\r\n' < "$project_root/.go-version")
output_dir="$project_root/.agents/skills/execute-test-cases/bin"

if [ -z "$version" ]; then
    echo "VERSION is empty" >&2
    exit 1
fi

actual_go_version=$(go env GOVERSION)
if [ "$actual_go_version" != "go$required_go_version" ]; then
    echo "Go $required_go_version is required, found $actual_go_version" >&2
    exit 1
fi

mkdir -p "$output_dir"
cd "$project_root"

build_target() {
    target_os=$1
    target_arch=$2
    suffix=$3
    output="$output_dir/ai-auto-test-store-$target_os-$target_arch$suffix"

    CGO_ENABLED=0 GOOS=$target_os GOARCH=$target_arch \
        go build -trimpath -buildvcs=false \
        -ldflags "-s -w -buildid= -X main.version=$version" \
        -o "$output" \
        ./cmd/ai-auto-test-store
    chmod 755 "$output"
}

build_target windows amd64 .exe
build_target windows arm64 .exe
build_target linux amd64 ''
build_target linux arm64 ''
build_target darwin amd64 ''
build_target darwin arm64 ''
