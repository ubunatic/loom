# SPDX-FileCopyrightText: 2026 Uwe Jugel
# SPDX-License-Identifier: AGPL-3.0-or-later

.PHONY: ⚙️ 🤖  # ⚙️ = manual/once, 🤖 = managed
_prim := \033[36m
_rst  := \033[0m


help: 🤖  # show this help
	@grep -E '^[a-zA-Z_-]+:.*[⚙🤖].*#+' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*#+ "}; {printf "    $(_prim)%-15s$(_rst) %s\n", $$1, $$2}'

build: ⚙️  ## compile the package
	go build ./...

test: ⚙️ validate-spec geometry-replay  ## validate specs, vet and run the test suite
	go vet ./...
	go test ./...

validate-spec: ⚙️  ## validate YAML specs against JSON Schema (Python jsonschema + PyYAML)
	python3 scripts/validate-spec.py

geometry-replay: ⚙️  ## verify saved ANSI replay bytes against geometry goldens
	python3 scripts/check-geometry-replay.py

tidy: ⚙️  ## sync go.mod/go.sum
	go mod tidy

vet: ⚙️  ## run go vet
	go vet ./...

install: ⚙️
	@echo "nothing to install, just run 'make build' or 'make test'"
