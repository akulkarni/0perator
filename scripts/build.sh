#!/bin/bash
set -e

# Build script for 0perator
# Builds multi-platform binaries and generates checksums

# Colors
ORANGE='\033[38;5;196m'
GRAY='\033[90m'
RESET='\033[0m'

# Configuration
BINARY_NAME="0perator"
BUILD_DIR="dist"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"

# Platforms to build
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "linux/386"
    "linux/arm"
    "windows/amd64"
)

echo ""
echo -e "${ORANGE}     ██████╗ ██████╗ ███████╗██████╗  █████╗ ████████╗ ██████╗ ██████╗ ${RESET}"
echo -e "${ORANGE}    ██╔═████╗██╔══██╗██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗${RESET}"
echo -e "${ORANGE}    ██║██╔██║██████╔╝█████╗  ██████╔╝███████║   ██║   ██║   ██║██████╔╝${RESET}"
echo -e "${ORANGE}    ████╔╝██║██╔═══╝ ██╔══╝  ██╔══██╗██╔══██║   ██║   ██║   ██║██╔══██╗${RESET}"
echo -e "${ORANGE}    ╚██████╔╝██║     ███████╗██║  ██║██║  ██║   ██║   ╚██████╔╝██║  ██║${RESET}"
echo -e "${ORANGE}     ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝${RESET}"
echo ""
echo -e "${ORANGE}               Infrastructure for AI native development${RESET}"
echo ""
echo "──────────────────────────────────────────────────────────────────────────"
echo ""
echo "  Building ${ORANGE}${BINARY_NAME}${RESET} ${VERSION}"
echo ""

# Clean and create build directory
echo "  Cleaning build directory..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS="/" read -r os arch <<< "$platform"

    output_name="$BINARY_NAME"
    if [ "$os" = "windows" ]; then
        output_name="${BINARY_NAME}.exe"
    fi

    echo "  Building ${ORANGE}${os}-${arch}${RESET}..."

    GOOS=$os GOARCH=$arch go build \
        -ldflags "-X main.Version=$VERSION -s -w" \
        -o "$BUILD_DIR/$output_name" \
        ./cmd/0perator-mcp

    # Create archive
    platform_dir="$BUILD_DIR/${BINARY_NAME}-${VERSION}-${os}-${arch}"
    mkdir -p "$platform_dir"
    mv "$BUILD_DIR/$output_name" "$platform_dir/"

    if [ "$os" = "windows" ]; then
        archive_name="${BINARY_NAME}-${VERSION}-${os}-${arch}.zip"
        (cd "$BUILD_DIR" && zip -q -r "$archive_name" "$(basename "$platform_dir")")
    else
        archive_name="${BINARY_NAME}-${VERSION}-${os}-${arch}.tar.gz"
        (cd "$BUILD_DIR" && tar -czf "$archive_name" "$(basename "$platform_dir")")
    fi

    # Clean up temp directory
    rm -rf "$platform_dir"

    echo "    ${ORANGE}✓${RESET} $archive_name"
done

# Generate checksums
echo ""
echo "  Generating checksums..."
(cd "$BUILD_DIR" && shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt)
echo "    ${ORANGE}✓${RESET} checksums.txt"

echo ""
echo "──────────────────────────────────────────────────────────────────────────"
echo "  ${ORANGE}✨ Build complete!${RESET}"
echo "──────────────────────────────────────────────────────────────────────────"
echo ""
echo "  Artifacts in: ${ORANGE}${BUILD_DIR}/${RESET}"
ls -lh "$BUILD_DIR" | tail -n +2 | awk '{printf "    %s %s\n", $9, $5}'
echo ""
