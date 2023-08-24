betula:
	export CGO_CFLAGS="-D_LARGEFILE64_SOURCE" CGO_ENABLED=1
	go build -o betula ./cmd/betula

debug-run: clean betula
	./betula db.betula

run-with-port: betula
	./betula -port 8081 db.betula

clean:
	rm -f betula

test: clean betula
	go test ./db
	go test ./feeds
	go test ./readpage
	sh test-web.sh
	killall betula
