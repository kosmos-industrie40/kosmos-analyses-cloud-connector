.PHONY: build clean test lint

build:
	go build -o connector ./

clean:
	${RM} connector

test:
	go test ./...

lint:
	golangci-lint run ./...
	go vet ./...

coverage:
	go test -covermode=count -coverprofile cov ./...
	go tool cover -html=cov -o coverage.html

race:
	go test -short -race ./...
