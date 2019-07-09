.PHONY : build run fresh test clean

BIN := ebs-snapshot-exporter

HASH := $(shell git rev-parse --short HEAD)
COMMIT_DATE := $(shell git show -s --format=%ci ${HASH})
BUILD_DATE := $(shell date '+%Y-%m-%d %H:%M:%S')
VERSION := ${HASH} (${COMMIT_DATE})


build:
	go build -o ${BIN} -ldflags="-X 'main.buildVersion=${VERSION}' -X 'main.buildDate=${BUILD_DATE}'"

run:
	./zed

fresh: clean build run

test:
	go test

clean:
	go clean
	- rm -f ${BIN}
