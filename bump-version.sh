#!/bin/bash

# Version bump script for Robin project
# Bumps the build number in VERSION file and updates version.go
set -e

VERSION_FILE="VERSION"
VERSION_GO_FILE="config/version.go"

# Check if VERSION file exists
if [ ! -f "$VERSION_FILE" ]; then
    echo "❌ Error: VERSION file not found"
    exit 1
fi

# Read current version
CURRENT_VERSION=$(cat "$VERSION_FILE")
echo "📋 Current version: $CURRENT_VERSION"

# Parse version (major.minor.build)
IFS='.' read -r MAJOR MINOR BUILD <<< "$CURRENT_VERSION"

# Increment build number
NEW_BUILD=$((BUILD + 1))
NEW_VERSION="${MAJOR}.${MINOR}.${NEW_BUILD}"

echo "🔄 Bumping version to: $NEW_VERSION"

# Update VERSION file
echo "$NEW_VERSION" > "$VERSION_FILE"

# Update version.go file
cat > "$VERSION_GO_FILE" << EOF
package config

// Version is the current version 
var Version = "$NEW_VERSION"
EOF

echo "✅ Version bumped successfully!"
echo "📦 New version: $NEW_VERSION"
echo "📄 Updated files: $VERSION_FILE, $VERSION_GO_FILE"
