betula:
	go build -o betula ./cmd/betula

debug-run:
	go build -o betula ./cmd/betula && 	./betula db.betula
