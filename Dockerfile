FROM alpine:latest AS deploy
RUN apk --no-cache add ca-certificates
COPY ebs-snapshot-exporter /
COPY config.sample  /etc/ebs-snapshot-exporter/config.toml
VOLUME ["/etc/ebs-snapshot-exporter"]
CMD ["./ebs-snapshot-exporter", "--config", "/etc/ebs-snapshot-exporter/config.toml"]  
