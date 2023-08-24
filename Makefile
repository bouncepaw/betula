betula:
	CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -o betula ./cmd/betula

debug-run: clean betula
	./betula db.betula

run-with-port: betula
	./betula -port 8081 db.betula

clean:
	rm -f betula