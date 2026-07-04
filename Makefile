# SPDX-FileCopyrightText: 2026 Uwe Jugel
# SPDX-License-Identifier: AGPL-3.0-or-later

.PHONY: ⚙️  # make all targets phony

help: ⚙️  ## show this help
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ⚙️  ## compile the package
	go build ./...

test: ⚙️  ## run the test suite
	go test ./...

tidy: ⚙️  ## sync go.mod/go.sum
	go mod tidy

vet: ⚙️  ## run go vet
	go vet ./...
