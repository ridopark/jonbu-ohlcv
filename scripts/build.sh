#!/bin/bash

# Build script for jonbu-ohlcv

set -e

echo "Building jonbu-ohlcv..."

# Create build directory
mkdir -p bin

# Build server
echo "Building server..."
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Build CLI
echo "Building CLI..."
go build -ldflags="-s -w" -o bin/cli cmd/cli/main.go

echo "Build completed successfully!"
echo "Binaries available in ./bin/"
