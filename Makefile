PROJECT_NAME := titea
PROJECT_ROOT := github.com/gengmei-tech/$(PROJECT_NAME)
FILES := $(shell find . -name "*.go" | grep -vE "vendor")
PACKAGES := $(shell go list ./...| grep -vE "vendor")

LDFLAGS += -X "$(PROJECT_ROOT)/server/store.BuildTs=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "$(PROJECT_ROOT)/server/store.GitHash=$(shell git rev-parse --short HEAD)"
LDFLAGS += -X "$(PROJECT_ROOT)/server/store.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "$(PROJECT_ROOT)/server/store.ReleaseVersion=$(shell git tag  --contains)"
LDFLAGS += -X "$(PROJECT_ROOT)/server/store.GoVersion=$(shell go version)"

.PHONY: all build check fmt lint

default: build

all: check build

build:
	go build -ldflags '$(LDFLAGS)' -o bin/titea ./cmd/server/*

check: fmt lint

fmt:
	@echo "gofmt"
	@ gofmt -s -l -w $(FILES) 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

lint:
	@echo "golint"
	@ golint -set_exit_status $(PACKAGES)

benchmark:
	go build -o bin/benchmark cmd/benchmark/main.go
