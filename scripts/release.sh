#!/bin/bash

# Release script for deeplx-cli
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v1.0.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version is provided
if [ $# -eq 0 ]; then
    print_error "Please provide a version number"
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format (should start with v)
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+.*$ ]]; then
    print_warning "Version should follow semantic versioning format (e.g., v1.0.0)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    print_error "Working directory is not clean. Please commit or stash your changes."
    exit 1
fi

# Check if tag already exists
if git tag -l | grep -q "^${VERSION}$"; then
    print_error "Tag ${VERSION} already exists"
    exit 1
fi

# Get current branch
CURRENT_BRANCH=$(git branch --show-current)
print_info "Current branch: ${CURRENT_BRANCH}"

# Confirm release
print_info "Preparing to release version: ${VERSION}"
print_info "This will:"
echo "  1. Run tests"
echo "  2. Create and push tag ${VERSION}"
echo "  3. Trigger GitHub Actions to build and release"
echo
read -p "Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_info "Release cancelled"
    exit 0
fi

# Run tests
print_info "Running tests..."
if ! make test; then
    print_error "Tests failed. Please fix them before releasing."
    exit 1
fi

# Run vet
print_info "Running go vet..."
if ! make vet; then
    print_error "go vet failed. Please fix the issues before releasing."
    exit 1
fi

# Test build
print_info "Testing build..."
if ! make build VERSION=${VERSION#v}; then
    print_error "Build failed. Please fix the issues before releasing."
    exit 1
fi

# Create and push tag
print_info "Creating tag ${VERSION}..."
git tag -a "${VERSION}" -m "Release ${VERSION}"

print_info "Pushing tag to origin..."
git push origin "${VERSION}"

print_info "✅ Release ${VERSION} has been created and pushed!"
print_info "GitHub Actions will now build and create the release automatically."
print_info "You can monitor the progress at: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"

# Clean up test build
rm -f deeplx-cli

print_info "🎉 Release process completed!"
