#!/bin/sh
# Build the browser version into ../docs/desk, which is what GitHub Pages serves.
#
# Both toolchains are carried: TinyGo at docs/desk/ because it is a quarter the
# size and is fetched before anything appears, and the standard Go build at
# docs/desk/go/ because TinyGo occasionally miscompiles something and having the
# other one a click away is how you find out that is what happened. The two
# pages link to each other.
#
#   ./build.sh          both
#   ./build.sh tinygo   TinyGo only
#   ./build.sh go       standard Go only
set -eu

cd "$(dirname "$0")"
out=../docs/desk

build_tinygo() {
	mkdir -p "$out"
	tinygo build -o "$out/pisano-desk.wasm" -target wasm -no-debug ./desk
	cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" "$out/wasm_exec.js"
}

build_go() {
	mkdir -p "$out/go"
	GOOS=js GOARCH=wasm go build -o "$out/go/pisano-desk.wasm" ./desk
	cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out/go/wasm_exec.js"
}

case "${1:-both}" in
both)   build_tinygo; build_go ;;
tinygo) build_tinygo ;;
go)     build_go ;;
*)      echo "usage: $0 [both|tinygo|go]" >&2; exit 2 ;;
esac

ls -lh "$out"/*.wasm "$out"/go/*.wasm 2>/dev/null || true
