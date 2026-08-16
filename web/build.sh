#!/bin/sh
# Build the browser version into ../docs/term, which is what GitHub Pages serves.
#
# TinyGo is the default because the binary is a quarter the size and this one is
# fetched over the network before anything appears. Pass "go" to build with the
# standard toolchain instead, into ../docs/term/go — worth having when TinyGo
# miscompiles something, which is the usual reason to want it.
set -eu

cd "$(dirname "$0")"
out=../docs/term

case "${1:-tinygo}" in
tinygo)
	mkdir -p "$out"
	tinygo build -o "$out/pisano.wasm" -target wasm -no-debug .
	cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" "$out/wasm_exec.js"
	;;
go)
	mkdir -p "$out/go"
	GOOS=js GOARCH=wasm go build -o "$out/go/pisano.wasm" .
	cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out/go/wasm_exec.js"
	;;
*)
	echo "usage: $0 [tinygo|go]" >&2
	exit 2
	;;
esac

ls -lh "$out"/*.wasm "$out"/go/*.wasm 2>/dev/null || true
