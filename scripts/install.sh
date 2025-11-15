#!/bin/sh
set -e

# 0perator installation script
# Based on the tiger-cli installer pattern
# Usage: curl -fsSL https://cli.0p.dev | sh

# Colors (using 256-color mode for compatibility)
ORANGE='\033[38;5;196m'
GRAY='\033[90m'
RESET='\033[0m'

# Configuration
REPO="akulkarni/0perator"
BINARY_NAME="0perator"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

# Platform detection
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Darwin)
            OS="darwin"
            ;;
        Linux)
            OS="linux"
            ;;
        MINGW* | MSYS* | CYGWIN*)
            OS="windows"
            ;;
        *)
            echo "${ORANGE}✗${RESET} Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64 | amd64)
            ARCH="amd64"
            ;;
        arm64 | aarch64)
            ARCH="arm64"
            ;;
        i386 | i686)
            ARCH="386"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            echo "${ORANGE}✗${RESET} Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Check dependencies
check_dependencies() {
    if ! command -v curl >/dev/null 2>&1; then
        echo "${ORANGE}✗${RESET} curl is required but not installed"
        exit 1
    fi

    if ! command -v tar >/dev/null 2>&1 && [ "$OS" != "windows" ]; then
        echo "${ORANGE}✗${RESET} tar is required but not installed"
        exit 1
    fi

    # Check for shasum or sha256sum
    if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
        echo "${GRAY}Warning: No SHA256 utility found, skipping checksum verification${RESET}"
        SKIP_CHECKSUM=1
    fi
}

# Get latest version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            echo "${ORANGE}✗${RESET} Failed to fetch latest version"
            exit 1
        fi
    fi
}

# Download with retry
download_with_retry() {
    url="$1"
    output="$2"
    max_attempts=3
    attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -fsSL "$url" -o "$output"; then
            return 0
        fi

        if [ $attempt -lt $max_attempts ]; then
            echo "${GRAY}Download failed, retrying ($attempt/$max_attempts)...${RESET}"
            sleep $((attempt * 2))
        fi

        attempt=$((attempt + 1))
    done

    return 1
}

# Install binary
install_binary() {
    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    echo ""
    echo "${ORANGE}     ██████╗ ██████╗ ███████╗██████╗  █████╗ ████████╗ ██████╗ ██████╗ ${RESET}"
    echo "${ORANGE}    ██╔═████╗██╔══██╗██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔═══██╗██╔══██╗${RESET}"
    echo "${ORANGE}    ██║██╔██║██████╔╝█████╗  ██████╔╝███████║   ██║   ██║   ██║██████╔╝${RESET}"
    echo "${ORANGE}    ████╔╝██║██╔═══╝ ██╔══╝  ██╔══██╗██╔══██║   ██║   ██║   ██║██╔══██╗${RESET}"
    echo "${ORANGE}    ╚██████╔╝██║     ███████╗██║  ██║██║  ██║   ██║   ╚██████╔╝██║  ██║${RESET}"
    echo "${ORANGE}     ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝${RESET}"
    echo ""
    echo "${ORANGE}               Infrastructure for AI native development${RESET}"
    echo ""
    echo "──────────────────────────────────────────────────────────────────────────"
    echo ""
    echo "  Installing ${ORANGE}0perator${RESET} ${VERSION} for ${PLATFORM}..."
    echo ""

    # Construct download URLs
    if [ "$OS" = "windows" ]; then
        ARCHIVE_NAME="${BINARY_NAME}-${VERSION}-${PLATFORM}.zip"
    else
        ARCHIVE_NAME="${BINARY_NAME}-${VERSION}-${PLATFORM}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/$REPO/releases/download/${VERSION}/${ARCHIVE_NAME}"
    CHECKSUM_URL="https://github.com/$REPO/releases/download/${VERSION}/checksums.txt"

    # Download binary archive
    echo "  Downloading binary..."
    if ! download_with_retry "$DOWNLOAD_URL" "$TMPDIR/$ARCHIVE_NAME"; then
        echo "${ORANGE}✗${RESET} Failed to download binary from $DOWNLOAD_URL"
        exit 1
    fi

    # Download and verify checksum
    if [ -z "$SKIP_CHECKSUM" ]; then
        echo "  Verifying checksum..."
        if download_with_retry "$CHECKSUM_URL" "$TMPDIR/checksums.txt"; then
            cd "$TMPDIR"

            if command -v shasum >/dev/null 2>&1; then
                if ! grep "$ARCHIVE_NAME" checksums.txt | shasum -a 256 -c --quiet 2>/dev/null; then
                    echo "${ORANGE}✗${RESET} Checksum verification failed"
                    exit 1
                fi
            elif command -v sha256sum >/dev/null 2>&1; then
                if ! grep "$ARCHIVE_NAME" checksums.txt | sha256sum -c --quiet 2>/dev/null; then
                    echo "${ORANGE}✗${RESET} Checksum verification failed"
                    exit 1
                fi
            fi
        else
            echo "${GRAY}Warning: Could not download checksums, skipping verification${RESET}"
        fi
    fi

    # Extract binary
    echo "  Extracting..."
    cd "$TMPDIR"
    if [ "$OS" = "windows" ]; then
        unzip -q "$ARCHIVE_NAME"
    else
        tar -xzf "$ARCHIVE_NAME"
    fi

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Install binary
    echo "  Installing to $INSTALL_DIR..."
    # Find the binary in the extracted directory
    EXTRACTED_DIR="${BINARY_NAME}-${VERSION}-${PLATFORM}"
    if [ "$OS" = "windows" ]; then
        mv "$EXTRACTED_DIR/${BINARY_NAME}.exe" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/${BINARY_NAME}.exe"
    else
        mv "$EXTRACTED_DIR/$BINARY_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Verify installation
    if [ "$OS" = "windows" ]; then
        INSTALLED_PATH="$INSTALL_DIR/${BINARY_NAME}.exe"
    else
        INSTALLED_PATH="$INSTALL_DIR/$BINARY_NAME"
    fi

    if [ ! -x "$INSTALLED_PATH" ]; then
        echo "${ORANGE}✗${RESET} Installation failed: binary not found or not executable"
        exit 1
    fi

    # Test the binary
    if ! "$INSTALLED_PATH" version >/dev/null 2>&1; then
        echo "${ORANGE}✗${RESET} Installation failed: binary not working"
        exit 1
    fi

    INSTALLED_VERSION=$("$INSTALLED_PATH" version)

    echo ""
    echo "  ${ORANGE}✓${RESET} Installed: $INSTALLED_VERSION"
    echo ""

    # Check if install directory is in PATH
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            ;;
        *)
            echo "──────────────────────────────────────────────────────────────────────────"
            echo ""
            echo "  ${GRAY}Note: $INSTALL_DIR is not in your PATH${RESET}"
            echo "  Add it to your shell configuration:"
            echo ""
            echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
            echo ""
            ;;
    esac

    echo "──────────────────────────────────────────────────────────────────────────"
    echo ""
    echo "  Next steps:"
    echo "    ${ORANGE}1.${RESET} Run ${ORANGE}0perator init${RESET} to configure your IDE"
    echo "    ${ORANGE}2.${RESET} Restart your IDE"
    echo "    ${ORANGE}3.${RESET} Try: \"Create a task management app\""
    echo ""
    echo "  Docs: ${ORANGE}https://0p.dev/docs${RESET}"
    echo ""
}

# Main
main() {
    detect_platform
    check_dependencies
    get_latest_version
    install_binary
}

main
