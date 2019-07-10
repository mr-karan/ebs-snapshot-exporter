FROM alpine:latest AS deploy
COPY ebs-snapshot-exporter /
COPY config.sample  /etc/ebs-snapshot-exporter/config.toml
VOLUME ["/etc/ebs-snapshot-exporter"]
CMD ["./ebs-snapshot-exporter", "--config", "/etc/ebs-snapshot-exporter/config.toml"]  
