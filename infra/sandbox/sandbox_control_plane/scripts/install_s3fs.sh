#!/bin/bash
# Install s3fs-fuse for S3 workspace mounting
#
# This script installs s3fs-fuse on the host system, which is required
# for mounting S3 buckets (including MinIO) to Docker containers.
#
# Usage:
#   sudo ./scripts/install_s3fs.sh

set -e

echo "=== Installing s3fs-fuse for S3 workspace mounting ==="

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Detected Linux system"

    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        echo "This script must be run as root (use sudo)"
        exit 1
    fi

    # Update package list
    echo "Updating package list..."
    apt-get update -qq

    # Install s3fs
    echo "Installing s3fs-fuse..."
    apt-get install -y s3fs

    # Verify installation
    if command -v s3fs &> /dev/null; then
        echo "✓ s3fs installed successfully: $(which s3fs)"
        s3fs --version
    else
        echo "✗ s3fs installation failed"
        exit 1
    fi

    # Check /dev/fuse device
    if [[ -e /dev/fuse ]]; then
        echo "✓ /dev/fuse device exists"
    else
        echo "✗ /dev/fuse device not found"
        echo "  You may need to load the fuse module: modprobe fuse"
    fi

elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "Detected macOS system"

    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        echo "Homebrew is not installed. Please install Homebrew first:"
        echo "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        exit 1
    fi

    # Install s3fs
    echo "Installing s3fs-fuse via Homebrew..."
    brew install s3fs-fuse

    # Verify installation
    if command -v s3fs &> /dev/null; then
        echo "✓ s3fs installed successfully: $(which s3fs)"
        s3fs --version
    else
        echo "✗ s3fs installation failed"
        exit 1
    fi

else
    echo "Unsupported OS: $OSTYPE"
    echo "Please install s3fs manually:"
    echo "  - Ubuntu/Debian: sudo apt-get install s3fs"
    echo "  - macOS: brew install s3fs-fuse"
    echo "  - CentOS/RHEL: sudo yum install s3fs-fuse"
    exit 1
fi

echo ""
echo "=== Installation complete ==="
echo ""
echo "Next steps:"
echo "1. Ensure executor Docker image includes s3fs (or mount it from host)"
echo "2. Start the control plane with S3 credentials configured"
echo "3. Create a session and verify S3 workspace is mounted"
echo ""
echo "Verification commands:"
echo "  # Check s3fs installation"
echo "  which s3fs"
echo ""
echo "  # Test S3 mount (optional)"
echo "  mkdir -p /tmp/test-s3"
echo "  s3fs your-bucket /tmp/test-s3 -o passwd_file=/etc/passwd-s3fs"
echo "  df -h | grep s3fs"
echo "  fusermount -u /tmp/test-s3"
echo ""
