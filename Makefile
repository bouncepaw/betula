export CGO_ENABLED=0

.PHONY: betula debug-run run-with-port clean test

betula:
	go build -o betula ./cmd/betula

debug-run: clean betula
	./betula db.betula

run-with-port: betula
	./betula -port 8081 db.betula

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
