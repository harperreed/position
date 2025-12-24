.PHONY: build test test-race test-coverage lint clean install coverage check

build:
	go build -o position ./cmd/position

test:
	go test -v ./...

test-race:
	go test -short -race -v ./...

test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

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
