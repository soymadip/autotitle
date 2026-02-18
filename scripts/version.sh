#!/usr/bin/env bash

# Use grep to get module name from go.mod to avoid dependency on 'go' command during early mise load
MODULE=$(grep "^module" go.mod | awk '{print $2}')
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

echo "MODULE=$MODULE"
echo "VERSION=$VERSION"
echo "COMMIT=$COMMIT"
echo "DATE=$DATE"
echo "LDFLAGS_BASE=-X \"$MODULE/internal/version.Version=$VERSION\" -X \"$MODULE/internal/version.Commit=$COMMIT\" -X \"$MODULE/internal/version.Date=$DATE\""
echo "LDFLAGS_RELEASE=-s -w -X \"$MODULE/internal/version.Version=$VERSION\" -X \"$MODULE/internal/version.Commit=$COMMIT\" -X \"$MODULE/internal/version.Date=$DATE\""
