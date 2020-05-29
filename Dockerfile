
FROM golang:1.14-buster AS builder

COPY . /go/src/gitlab.inovex.io/proj-kosmos/kosmos-analyse-cloud-connector
RUN cd /go/src/gitlab.inovex.io/proj-kosmos/kosmos-analyse-cloud-connector; go build -o /usr/local/bin/connector

FROM debian:10-slim

COPY --from=builder /usr/local/bin/connector /usr/local/bin/connector

RUN apt update && apt upgrade -y
RUN adduser --system --home /home/kosmos kosmos

USER kosmos:nogroup

ENTRYPOINT ["/usr/local/bin/connector"]
