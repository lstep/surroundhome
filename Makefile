.PHONY: all build test coverage clean

all: build

build:
	cd surserver/cmd/surserver && go build -o ../../../bin/surserver


test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf bin
	rm -f coverage.out
