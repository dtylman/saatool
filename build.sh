#!/bin/bash

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -d "cmd" ]]; then
    print_error "This script must be run from the saatool root directory"
    exit 1
fi

# Set Android environment variables if not already set
if [[ -z "$ANDROID_HOME" ]]; then
    export ANDROID_HOME="$HOME/Android/Sdk"
    print_status "ANDROID_HOME not set, defaulting to $ANDROID_HOME"
fi
if [[ -z "$ANDROID_NDK_HOME" ]]; then
    export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/29.0.13846066"
    print_status "ANDROID_NDK_HOME not set, defaulting to $ANDROID_NDK_HOME"
fi
export PATH="$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools"

# Add Android NDK to PATH
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools

print_status "Starting SAATool build process..."

# Create build output directory
BUILD_DIR="./build"
mkdir -p "$BUILD_DIR"

# Read version from VERSION file
if [[ -f "VERSION" ]]; then
    PROJECT_VERSION=$(cat VERSION | tr -d '\n\r ')
    print_status "Using version from VERSION file: $PROJECT_VERSION"
else
    PROJECT_VERSION="unknown"
    print_warning "VERSION file not found, using 'unknown' as version"
fi

print_status "Build directory: $BUILD_DIR"
print_status "Project version: $PROJECT_VERSION"

# Build CLI tool (saatooltool)
print_status "Building CLI tool (saatooltool)..."
cd cmd/saatooltool

if go build -o "../../$BUILD_DIR/saatooltool"; then
    print_success "CLI tool built successfully"
else
    print_error "Failed to build CLI tool"
    exit 1
fi

cd ../../

# Build desktop version for testing
print_status "Building desktop version for testing..."
cd cmd/saatool

if go build -o "../../$BUILD_DIR/saatool-desktop"; then
    print_success "Desktop version built successfully"
else
    print_error "Failed to build desktop version"
    exit 1
fi

cd ../../

# Android App Build
print_status "Building Android app..."

SOURCE_DIR="$(pwd)/cmd/saatool"
APP_ICON="$SOURCE_DIR/icon.png"
FYNE_METADATA="$SOURCE_DIR/FyneApp.toml"

# Check if required files exist
if [[ ! -f "$FYNE_METADATA" ]]; then
    print_error "FyneApp.toml not found at $FYNE_METADATA"
    exit 1
fi

if [[ ! -f "$APP_ICON" ]]; then
    print_error "App icon not found at $APP_ICON"
    exit 1
fi

# Extract metadata from FyneApp.toml
APP_ID=$(grep 'ID = ' "$FYNE_METADATA" | sed 's/.*ID = "\(.*\)".*/\1/')
APP_NAME=$(grep 'Name = ' "$FYNE_METADATA" | sed 's/.*Name = "\(.*\)".*/\1/')

# Use PROJECT_VERSION as APP_VERSION
APP_VERSION="$PROJECT_VERSION"

print_status "App metadata:"
print_status "  Name: $APP_NAME"
print_status "  Version: $APP_VERSION"
print_status "  ID: $APP_ID"

# Check if fyne command is available
if ! command -v fyne &> /dev/null; then
    print_warning "fyne command not found in PATH, trying ~/go/bin/fyne"
    FYNE_CMD="$HOME/go/bin/fyne"
    if [[ ! -f "$FYNE_CMD" ]]; then
        print_error "fyne command not found. Please install it with: go install fyne.io/tools/cmd/fyne@latest"
        exit 1
    fi
else
    FYNE_CMD="fyne"
fi

# Function to build Android APK
build_android_apk() {
    local target=$1
    local arch=$2
    local output_name="$APP_NAME-$arch.apk"
    local output_path="$BUILD_DIR/$output_name"
    
    print_status "Building $arch APK..."
    
    if $FYNE_CMD package --target "$target" --app-id "$APP_ID" --source-dir "$SOURCE_DIR" --name "$APP_NAME" --icon "$APP_ICON" --app-version="$APP_VERSION"; then
        # Move the generated APK to build directory with proper name
        if [[ -f "$SOURCE_DIR/$APP_NAME.apk" ]]; then
            mv "$SOURCE_DIR/$APP_NAME.apk" "$output_path"
            print_success "$arch APK built: $output_path"
        else
            print_error "APK file not found after build: $SOURCE_DIR/$APP_NAME.apk"
            return 1
        fi
    else
        print_error "Failed to build $arch APK"
        return 1
    fi
}

# Build ARM64 APK (for real devices)
build_android_apk "android/arm64" "arm64"

# Build AMD64 APK (for emulator)
build_android_apk "android/amd64" "amd64"

# Create symlinks for easier access to latest builds
print_status "Creating convenience symlinks..."

ln -sf "saatooltool" "$BUILD_DIR/saatooltool-latest"
ln -sf "saatool-desktop" "$BUILD_DIR/saatool-desktop-latest"
ln -sf "$APP_NAME-arm64.apk" "$BUILD_DIR/saatool-latest-arm64.apk"
ln -sf "$APP_NAME-amd64.apk" "$BUILD_DIR/saatool-latest-amd64.apk"

print_success "Build completed successfully!"
print_status "Build artifacts:"
print_status "  CLI tool: $BUILD_DIR/saatooltool"
print_status "  Desktop: $BUILD_DIR/saatool-desktop"
print_status "  Android ARM64: $BUILD_DIR/$APP_NAME-arm64.apk"
print_status "  Android AMD64: $BUILD_DIR/$APP_NAME-amd64.apk"
print_status ""
print_status "Latest build symlinks created in $BUILD_DIR/"
print_status "Use 'ls -la $BUILD_DIR/' to see all build artifacts"

# Optional: Test CLI tool
if [[ "$1" == "--test" ]]; then
    print_status "Testing CLI tool..."
    if "$BUILD_DIR/saatooltool" --help > /dev/null 2>&1; then
        print_success "CLI tool test passed"
    else
        print_warning "CLI tool test failed - binary may not be compatible with current system"
    fi
fi

print_success "All builds completed successfully!"