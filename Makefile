PACKAGE = mux
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell echo $(shell cat $(PWD)/.version)-$(shell git describe --tags --always))

GODOC   = godoc
GOFMT   = gofmt

ifneq ($(wildcard /snap/go/current/bin/go),)
	GO = /snap/go/current/bin/go
else ifneq ($(shell which go1.11),)
	GO = go1.11
else
	GO = go
endif

ifneq ($(wildcard ./bin/golangci-lint),)
	GOLINT = ./bin/golangci-lint
else
	GOLINT = golangci-lint
endif

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[0;35m▶\033[0m")

.PHONY: all
all: check

# Vendor
.PHONY: vendor
vendor:
	$(info $(M) running go mod vendor…) @
	$Q $(GO) mod vendor

# Tidy
.PHONY: tidy
tidy:
	$(info $(M) running go mod tidy…) @
	$Q $(GO) mod tidy

# Check
.PHONY: check
check: vendor lint test

# Lint
.PHONY: lint
lint:
	$(info $(M) running $(GOLINT)…)
	$Q $(GOLINT) run

# Test
.PHONY: test
test:
	$(info $(M) running go test…) @
	$Q $(GO) test -cover -race -v ./...

.PHONY: fmt
fmt:
	$(info $(M) running $(GOFMT)…) @
	$Q $(GOFMT) ./...

.PHONY: doc
doc:
	$(info $(M) running $(GODOC)…) @
	$Q $(GODOC) ./...

.PHONY: version
version:
	@echo $(VERSION)
