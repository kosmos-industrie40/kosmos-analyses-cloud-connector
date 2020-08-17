FROM golang:1.14-buster AS builder
COPY . /go/src/gitlab.inovex.io/proj-kosmos/kosmos-analyse-cloud-connector
WORKDIR /go/src/gitlab.inovex.io/proj-kosmos/kosmos-analyse-cloud-connector
RUN go build -o /usr/local/bin/connector

FROM gcr.io/distroless/base-debian10:latest
COPY --from=builder /usr/local/bin/connector /usr/local/bin/connector
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/connector"]
