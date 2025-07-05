#!/usr/bin/env bash

PACKAGE="github.com/nekr0z/gk/internal/version"
VERSION=$(git describe --always --dirty)
DATE=$(date --rfc-3339=seconds)

LDFLAGS=(
  "-X '${PACKAGE}.buildVersion=${VERSION}'"
  "-X '${PACKAGE}.buildDate=${DATE}'"
)

go build -ldflags="${LDFLAGS[*]}" -o gk ./cmd/gk/main.go
