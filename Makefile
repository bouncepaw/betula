# SPDX-FileCopyrightText: 2022-2026 Betula contributors
#
# SPDX-License-Identifier: AGPL-3.0-only

export CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
export CGO_ENABLED=1

.PHONY: betula debug-run run-with-port clean test

betula:
	go build -o betula ./cmd/betula

debug-run: clean betula
	./betula db.betula

run-with-port: betula
	./betula -port 8081 db.betula

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

clean:
	rm -f betula

test: clean betula
	go test ./db
	go test ./types
	go test ./feeds
	go test ./readpage
	go test ./fediverse/activities
	sh test-web.sh
	killall betula
