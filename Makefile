ifeq (,$(wildcard .env))	# if .env file does not exist then copy env.example to .env
	cp env.example .env
endif

.PHONY: run

run: build-test
	@CGO_ENABLED=1 ./bin/dh

build-test:
	@go build -o bin/dh main.go

build: build-linux build-windows

build-linux:
	@echo "Building for linux"
	@GOOS=linux GOARCH=amd64 go build -o bin/dh-linux-amd64 main.go
	@GOOS=linux GOARCH=arm64 go build -o bin/dh-linux-arm64 main.go

build-windows:
	@echo "Building for windows"
	@GOOS=windows GOARCH=amd64 go build -o bin/dh-win-amd64.exe main.go
	@GOOS=windows GOARCH=arm64 go build -o bin/dh-win-arm64.exe main.go

test:
	@go test -v ./...