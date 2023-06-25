betula:
	go build -o betula ./cmd/betula

debug-run: clean betula
	./betula db.betula

run-with-port: betula
	./betula -port 8081 db.betula

clean:
	rm -f betula