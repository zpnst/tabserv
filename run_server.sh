#!/bin/sh

BINARY="bin/tabserv-server"

echo "[*] building server..."
mkdir -p bin
go build -o "$BINARY" ./cmd/server

echo "[*] running server as root..."
exec sudo "$BINARY" "$@"