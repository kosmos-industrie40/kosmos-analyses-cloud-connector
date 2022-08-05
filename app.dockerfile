FROM golang:1.14-buster AS builder
COPY . /go/src/github.com/kosmos-industrie40/kosmos-analyses-cloud-connector
WORKDIR /go/src/github.com/kosmos-industrie40/kosmos-analyses-cloud-connector
RUN go build -o /usr/local/bin/connector src/main.go

FROM gcr.io/distroless/base-debian10:latest
COPY --from=builder /usr/local/bin/connector /usr/local/bin/connector
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/connector"]
