# SPDX-FileCopyrightText: 2022-2026 Betula contributors
#
# SPDX-License-Identifier: AGPL-3.0-only

export CGO_ENABLED=0

ALL_FILES := $(shell find . -type f -name '*.go')

.PHONY: betula debug-run run-with-port clean test lint lint-fix crosscompile

betula: $(ALL_FILES)
	go build -o betula ./cmd/betula

crosscompile: dst/linux-arm64/betula dst/linux-amd64/betula dst/darwin-arm64/betula dst/darwin-amd64/betula

debug-run: clean betula
	./betula db.betula

run-with-port: betula
	./betula -port 8081 db.betula

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

clean:
	rm -rf betula dst

test: clean betula
	go test ./db
	go test ./types
	go test ./svc/feeds
	go test ./pkg/httpsig
	go test ./pkg/ticks
	go test ./web
	go test ./gateways/www
	go test ./fediverse/activities
	sh test-web.sh
	killall betula

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
