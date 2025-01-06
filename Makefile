format:
	gofumpt -l -w .

build:
	make format
	go build -ldflags '-w -s' -trimpath github.com/sqkam/geminipro
	docker build -t sqkam/geminipro:latest .

run:
	make build
	./geminipro
