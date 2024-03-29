.PHONY: build clean test lint

build:
	go build -o connector src/main.go

clean:
	${RM} connector

test:
	go test ./...

lint:
	golangci-lint run ./...

coverage:
	go test -covermode=count -coverprofile cov --tags unit ./...
	go tool cover -html=cov -o coverage.html

race:
	go test -short -race ./...

docker:
	docker build -t edge-cloud-connector -f app.dockerfile .
	docker build -t edge-cloud-connector-swagger -f swagger.dockerfile .
