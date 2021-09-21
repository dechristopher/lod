#!/bin/sh

if [ $# -gt 0 ]; then
  VERSION=$1
  echo "Building LOD v$VERSION for release..."

  echo "Building Darwin (amd64)"
  env GOOS=darwin GOARCH=amd64 go build -ldflags "-w -X github.com/tile-fund/lod/config.Version=${VERSION}" -o lod
  zip -9 build/artifacts/lod_${VERSION}_darwin_amd64.zip lod
  rm -rf lod

  echo "Building Linux (amd64)"
  env GOOS=linux GOARCH=amd64 go build -ldflags "-w -X github.com/tile-fund/lod/config.Version=${VERSION}" -o lod
  zip -9 build/artifacts/lod_${VERSION}_linux_amd64.zip lod
  rm -rf lod

  echo "Building Windows (amd64)"
  env GOOS=windows GOARCH=amd64 go build -ldflags "-w -X github.com/tile-fund/lod/config.Version=${VERSION}" -o lod
  zip -9 build/artifacts/lod_${VERSION}_windows_amd64.zip lod
  rm -rf lod

else
  echo "Must specify version as second argument!"
  exit 1
fi
