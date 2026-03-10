#!/bin/bash

# src/cz/bump-flake.sh - Commitizen hook to update flake.nix version

set -e

NEW_VERSION="${1:-$CZ_PRE_NEW_VERSION}"

# Check if version was passed
if [ -z "$NEW_VERSION" ]; then
    echo "Error: No version argument provided."
    exit 1
fi

# Remove leading 'v' if present for the version variable in flake.nix
CLEAN_VERSION="${NEW_VERSION#v}"

echo -e "\033[38;5;244m -> Bumping flake.nix to ${CLEAN_VERSION}...\033[0m"

# 1. Update version variable
sed -i "s/version = \".*\"; # Should be synced/version = \"${CLEAN_VERSION}\"; # Should be synced/" flake.nix

# 2. Update version in ldflags
sed -i "s/Version=v.*/Version=v${CLEAN_VERSION}\"/" flake.nix

# 3. Stage the file for the release commit
git add flake.nix

echo -e "\033[0;32m -> flake.nix version updated and staged.\033[0m"
