.PHONY: test test-race test-verbose bench vet fmt tidy build clean

test:
	go test ./...

test-race:
	go test -race ./...

test-verbose:
	go test -v ./...

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

build:
	go build ./...

clean:
	go clean ./...
