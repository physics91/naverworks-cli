VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/physics91/naverworks-cli/cmd.version=$(VERSION) \
	-X github.com/physics91/naverworks-cli/cmd.commit=$(COMMIT) \
	-X github.com/physics91/naverworks-cli/cmd.buildDate=$(BUILD_DATE)

.PHONY: build test test-fast test-full test-canary clean

build:
	go build -ldflags "$(LDFLAGS)" -o naverworks .

test:
	go test ./... -v

test-fast:
	go test ./internal/testkit/cli -run 'TestHarness|TestHarnessFailureCategories' -v -count=1
	go test ./internal/api ./internal/auth -v -count=1
	go test ./cmd -run 'Test(CommandMetaContracts|SharedCLIRunnerVersion|Smoke_|Journey(AuthStatus|ConfigLifecycle|DirectoryListUsers|BotSendText))' -v -count=1

test-full:
	go test ./... -v -count=1

test-canary:
	go test ./cmd -run TestBinaryCanaryVersion -v -count=1

clean:
	rm -f naverworks
