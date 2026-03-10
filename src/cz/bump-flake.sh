#!/bin/bash

# src/cz/bump-flake.sh - Commitizen hook to update flake.nix version

set -e

NEW_VERSION="${1:-$CZ_PRE_NEW_VERSION}"

# Check if version was passed
if [ -z "$NEW_VERSION" ]; then
    echo -e "\033[0;31m Error: No version argument provided.\033[0m"
    exit 1
fi

# Remove leading 'v' if present for the version variable in flake.nix
CLEAN_VERSION="${NEW_VERSION#v}"

echo -e "\033[38;5;244m -> Bumping flake.nix to ${CLEAN_VERSION}...\033[0m"

# 1. Update version variable
sed -i "s/\(version = \)\".*\"\(; # Should be synced\)/\1\"${CLEAN_VERSION}\"\2/" flake.nix

# 2. Update version in ldflags
sed -i "s/\(Version=v\)[^\"]*/\1${CLEAN_VERSION}/" flake.nix

# 3. Handle vendorHash update
# We attempt to build to see if the hash needs updating. 
# This is useful if dependencies changed between releases.
echo -e "\033[38;5;244m -> Checking vendorHash in flake.nix...\033[0m"

# Temporarily disable pipefail to catch the error from nix build
set +e
BUILD_OUTPUT=$(nix build . --no-link --print-out-paths --extra-experimental-features 'nix-command flakes' 2>&1)
set -e

if echo "$BUILD_OUTPUT" | grep -q "got:"; then
    NEW_HASH=$(echo "$BUILD_OUTPUT" | grep "got:" | sed 's/.*got: *//' | tr -d '[:space:]')
    if [ -n "$NEW_HASH" ]; then
        echo -e "\033[38;5;244m -> Updating vendorHash to ${NEW_HASH}...\033[0m"
        sed -i "s/\(vendorHash = \)\".*\"\(;\)/\1\"${NEW_HASH}\"\2/" flake.nix
    fi
fi

# 4. Stage the file for the release commit
git add flake.nix

echo -e "\033[0;32m -> flake.nix version updated and staged.\033[0m"
