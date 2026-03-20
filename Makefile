VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/physics91/naverworks-cli/cmd.version=$(VERSION) \
	-X github.com/physics91/naverworks-cli/cmd.commit=$(COMMIT) \
	-X github.com/physics91/naverworks-cli/cmd.buildDate=$(BUILD_DATE)

.PHONY: build test clean

build:
	go build -ldflags "$(LDFLAGS)" -o nw-cli .

test:
	go test ./... -v

clean:
	rm -f nw-cli
