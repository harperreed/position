.PHONY: build test test-race lint clean install coverage check

build:
	go build -o position ./cmd/position

test:
	go test -v ./...

test-race:
	go test -race -v ./...

lint:
	golangci-lint run

clean:
	rm -f position
	rm -f coverage.out

install:
	go install ./cmd/position

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

check: lint test-race
