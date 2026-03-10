#!/bin/bash

# src/cz/bump-pkgbuild.sh - Commitizen hook to update PKGBUILD version

set -e

NEW_VERSION="${1:-$CZ_PRE_NEW_VERSION}"

# Check if version was passed
if [ -z "$NEW_VERSION" ]; then
    echo -e "\033[0;31m Error: No version argument provided.\033[0m"
    exit 1
fi

# Remove leading 'v' if present
CLEAN_VERSION="${NEW_VERSION#v}"

echo -e "\033[38;5;244m -> Bumping PKGBUILD to ${CLEAN_VERSION}...\033[0m"

# 1. Update pkgver (match only at the start of the line or with whitespace)
sed -i "s/^\( *pkgver=\).*/\1${CLEAN_VERSION}/" PKGBUILD

# 2. Reset pkgrel to 1
sed -i "s/^\( *pkgrel=\).*/\11/" PKGBUILD

# 3. Stage the file for the release commit
git add PKGBUILD

echo -e "\033[0;32m -> PKGBUILD updated and staged.\033[0m"
