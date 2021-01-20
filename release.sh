#!/bin/bash

set -Eeuo pipefail

NAME="kci"
GIT_SHA="$(git --no-pager describe --always --dirty)"
BUILD_TIME="$(date '+%s')"
LFLAGS="-X main.gitsha=$GIT_SHA -X main.compiled=$BUILD_TIME"

VERSION="${VERSION:-$GIT_SHA}"
echo "Version: $VERSION"

release() {
  echo "Building $NAME-$VERSION for $GOOS/$GOARCH (${GOARM:-default})..."

  if [ "$GOOS" == "windows" ]; then
    EXT=".exe"
  else
    EXT=""
  fi

  if [ "$GOARCH" == "arm" -a -n "${GOARM:-}" ]; then
    ARM_EXT="-armv$GOARM"
  else
    ARM_EXT=""
  fi

  export GOARCH GOOS
  if [ -n "${GOARM:-}" ]; then
    export GOARM
  fi

  go generate
  CGO_ENABLED=0 go build -ldflags " -w $LFLAGS" -o "bin/$NAME$EXT"
  tar -czf "release/$NAME-$GOOS-$GOARCH$ARM_EXT.tar.gz" -C bin/ "$NAME$EXT"
  (cd release && sha1sum "$NAME-$GOOS-$GOARCH$ARM_EXT.tar.gz" > "$NAME-$GOOS-$GOARCH$ARM_EXT.tar.gz.sha1")
  rm -f "bin/$NAME$EXT"
}

rm -rf bin release
mkdir -p bin release

while read configuration; do
  unset GOOS
  unset GOARCH
  unset GOARM
  eval "$configuration"
  release
done <<EOF
GOOS=linux GOARCH=amd64
GOOS=linux GOARCH=arm64
GOOS=linux GOARCH=arm GOARM=5
GOOS=darwin GOARCH=amd64
EOF

CURRENT_TAG="$(git describe --tags --exact-match 2>/dev/null || true)"
CURRENT_COMMIT_ID="$(git rev-parse HEAD)"

if [ -z "$CURRENT_TAG" ]; then
  echo "Currently not on a tag. Skipping GitHub release creation..."
  exit 0
fi

gh auth status
if ! gh release list | cut -f3 | grep -qx "$VERSION"; then
  echo "Creating the $VERSION release..."
  gh release create "$VERSION" release/* -d --target "$CURRENT_COMMIT_ID" -n "Released v$VERSION" --title "v$VERSION"
else
  gh release upload "$VERSION" --clobber release/*
fi
