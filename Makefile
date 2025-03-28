export

##### Go variables #####

GOBIN := $(PWD)/bin

GOPATH := $(shell go env GOPATH)

# To build using a specific version of Go, specify the version on the command
# line.
#    GOVERSION=1.21
GOVERSION ?= $(shell go list -m -f '{{.GoVersion}}' | sort -V | tail -n 1)


# The length of the hash `git rev-parse` returns with the `--short` option
# without an explicit length is determined by the configuration of the host.
# We fix it to 7 here to make sure we get the same hash across different
# environments.
# Reference: https://git-scm.com/docs/git-rev-parse#Documentation/git-rev-parse.txt---shortlength
GIT_REF := $(shell git rev-parse --short=7 HEAD)

PATH := $(GOBIN):$(PATH)

SHELL := env "PATH=$(PATH)" bash

########## Dependencies ##########

.PHONY: tidy
tidy: ## Tidies Go modules.
	@go mod tidy -v

########## Build ##########

.PHONY: build
build: ## Builds binaries under //cmd into //bin.
	@go build \
		-o ./bin/ \
		./cmd/*

########## Test ##########

.PHONY: coverage
coverage: ## Tests all Go packages and writes a coverage profile to //coverage.out
	@make test args='-covermode=atomic -coverpkg=./... -coverprofile=coverage.out'

.PHONY: test
test: ## Tests all Go packages, including race detection.
	@go test $(args) -race -cover ./...

.PHONY: test/long
test/long: ## Tests all Go packages multiple times, including race detection and coverage summaries.
	@make test args='-covermode=atomic -coverpkg=./... -count=5'

.PHONY: test/v
test/v: ## Tests all Go packages, with verbose output.
	@make test args='-v'

########## Format ##########

.PHONY: fmt
fmt: ## Formats all Go files.
	@find . -iname "*.go" -not -path "./vendor/**" | xargs gofmt -s -w

########## Lint ##########

.PHONY: lint
lint: ## Lints all Go files.
	@golangci-lint run $(args) ./...

.PHONY: lint/fix
lint/fix: ## Attempts to fix lint errors in all Go files.
	@make lint args='--fix -v' cons_args='-v'

########## Help ##########

.PHONY: help
help: ## Displays a help message.
	@printf "Usage:\n  make \033[33m<target>\033[0m\n\n"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_0-9\/%_-]+:.*?##/ { printf "  \033[32m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
