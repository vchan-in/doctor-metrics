ifeq (,$(wildcard .env))	# if .env file does not exist then copy env.example to .env
	cp env.example .env
endif

.PHONY: run

run: build
	@CGO_ENABLED=1 ./bin/dh

build: 
	@go build -o bin/dh main.go

# test:
# 	@go test ./... -cover