ARG GO_VERSION=1.12
FROM golang:${GO_VERSION}-alpine AS builder
RUN apk update && apk add gcc libc-dev make git
WORKDIR /ebs-snapshot-exporter/
COPY ./ ./
ENV CGO_ENABLED=0 GOOS=linux
RUN make build

FROM alpine:latest AS deploy
RUN apk --no-cache add ca-certificates
WORKDIR /ebs-snapshot-exporter/
COPY --from=builder /ebs-snapshot-exporter/ ./
RUN mkdir -p /etc/ebs-snapshot-exporter && cp config.sample /etc/ebs-snapshot-exporter/config.toml
# Define data volumes
VOLUME ["/etc/ebs-snapshot-exporter"]
CMD ["./ebs-snapshot-exporter", "--config", "/etc/ebs-snapshot-exporter/config.toml"]  
