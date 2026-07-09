# SPDX-FileCopyrightText: 2022-2026 Betula contributors
#
# SPDX-License-Identifier: AGPL-3.0-only

export CGO_ENABLED=0

ALL_FILES := $(shell find . -type f -name '*.go*') web/style.css
HELP_SPACING := 25

.PHONY: help betula debug-run clean test lint lint-fix crosscompile

betula: $(ALL_FILES) ## Build the betula binary
	go build -o betula ./cmd/betula

crosscompile: dst/linux-arm64/betula dst/linux-amd64/betula dst/darwin-arm64/betula dst/darwin-amd64/betula ## Cross-compile the betula binary for multiple platforms

debug-run: clean betula ## Run the betula binary
	./betula db.betula

lint: ## Run the linter
	golangci-lint run

lint-fix: ## Fix lint issues
	golangci-lint run --fix

clean: ## Clean up build artifacts
	rm -rf betula dst

test: clean betula ## Run tests
	go test ./db
	go test ./types
	go test ./svc/feeds
	go test ./pkg/httpsig
	go test ./pkg/bxtime
	go test ./web
	go test ./gateways/www
	go test ./fediverse/activities
	sh test-web.sh
	killall betula

# SPDX-SnippetBegin
# SPDX-License-Identifier: GPL-3.0-only
# SPDX-SnippetCopyrightText: 2026 A Possible Space Ltd. <us@possible.space>
help: ## Display this help message
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | sort | awk -F ':.*?## ' 'NF==2 {printf "  \033[36m%-$(HELP_SPACING)s\033[0m %s\n", $$1, $$2}'
# SPDX-SnippetEnd

dst:
	mkdir -p dst/linux-arm64 dst/linux-amd64 dst/darwin-arm64 dst/darwin-amd64

dst/linux-arm64/betula: dst $(ALL_FILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dst/linux-arm64/betula ./cmd/betula
dst/linux-amd64/betula: dst $(ALL_FILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dst/linux-amd64/betula ./cmd/betula
dst/darwin-arm64/betula: dst $(ALL_FILES)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dst/darwin-arm64/betula ./cmd/betula
dst/darwin-amd64/betula: dst $(ALL_FILES)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dst/darwin-amd64/betula ./cmd/betula

INKSCAPE ?= /Applications/Inkscape.app/Contents/MacOS/inkscape --export-area-snap --export-png-antialias=1
web/pix/favicon.png: web/pix/favicon.svg
	$(INKSCAPE) -w 32 -h 32 $< -o $@
web/pix/icon-192.png: web/pix/favicon.svg
	$(INKSCAPE) -w 192 -h 192 $< -o $@
web/pix/icon-512.png: web/pix/favicon.svg
	$(INKSCAPE) -w 512 -h 512 $< -o $@

