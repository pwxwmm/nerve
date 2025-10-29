#!/bin/bash
# Nerve Agent Build Script
#
# Author: mmwei3 (2025-10-28)
# Wethers: cloudWays
#
# This script builds the Nerve Agent binary

set -e

echo "Building Nerve Agent..."

cd "$(dirname "$0")/../agent"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

# Build the agent
echo "Compiling agent binary..."
go build -o nerve-agent .

# Check if build succeeded
if [ -f "nerve-agent" ]; then
    echo "✓ Agent binary built successfully: agent/nerve-agent"
    echo "Binary size: $(du -h nerve-agent | cut -f1)"
else
    echo "✗ Build failed: nerve-agent not found"
    exit 1
fi

