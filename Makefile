.PHONY: all build test coverage clean plugins

all: build plugins

build:
	cd surserver/cmd/surserver && go build -o ../../../dist/surserver
	#-ldflags -X main.Version=$(git describe --tags --dirty --always --abbrev=12)
	@mkdir -p dist
	@cp surserver/cmd/surserver/config.yaml dist/

plugins:
	@echo "Building plugins..."
	@mkdir -p dist
	@for plugin in $$(find plugins -mindepth 1 -maxdepth 1 -type d -printf "%f\n"); do \
		echo "Building plugin: $$plugin"; \
		cd plugins/$$plugin && go build -o ../../dist/$$plugin; \
		if [ -f $$plugin-config.yaml ]; then \
			cp $$plugin-config.yaml ../../dist/; \
		fi; \
		cd ../..; \
	done

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -rf dist
	rm -f coverage.out
