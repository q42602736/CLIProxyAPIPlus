#!/usr/bin/env bash
#
# build.sh - Cross-platform Build Script
# Builds CLIProxyAPIPlus for macOS ARM64 and Windows AMD64
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== CLIProxyAPIPlus Cross-Platform Build ===${NC}"
echo ""

# Get version information
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "Build Information:"
echo "  Version:    ${VERSION}-plus"
echo "  Commit:     ${COMMIT}"
echo "  Build Date: ${BUILD_DATE}"
echo ""

# Build flags
LDFLAGS="-s -w -X 'main.Version=${VERSION}-plus' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'"

# Build function
build() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT_NAME=$3

    echo -e "${YELLOW}Building ${GOOS}/${GOARCH}...${NC}"

    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${OUTPUT_NAME}" \
        ./cmd/server/

    if [ $? -eq 0 ]; then
        local SIZE=$(du -h "${OUTPUT_NAME}" | cut -f1)
        echo -e "${GREEN}✓ Built ${OUTPUT_NAME} (${SIZE})${NC}"
    else
        echo -e "${RED}✗ Failed to build ${OUTPUT_NAME}${NC}"
        exit 1
    fi
    echo ""
}

# Build for macOS ARM64
echo -e "${GREEN}Building for macOS ARM64...${NC}"
build "darwin" "arm64" "CLIProxyAPIPlus"

# Build for Windows AMD64
echo -e "${GREEN}Building for Windows AMD64...${NC}"
build "windows" "amd64" "CLIProxyAPIPlus.exe"

echo -e "${GREEN}=== Build Complete ===${NC}"
echo ""
echo "Built binaries in current directory:"
ls -lh CLIProxyAPIPlus CLIProxyAPIPlus.exe 2>/dev/null || true
echo ""
echo -e "${YELLOW}Tip: You can run ./CLIProxyAPIPlus directly or use pm2 to manage it.${NC}"
