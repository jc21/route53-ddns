# Go parameters
VERSION=$(shell cat package.json | jq -r .version)
GITCOMMIT=$(shell git rev-list -1 HEAD)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=route53-ddns

all: build

build:
	GO111MODULE=on $(GOBUILD) -o bin/$(BINARY_NAME) -v -ldflags "-X config.gitCommit=$(GITCOMMIT) -X config.appVersion=$(VERSION)" ./cmd/route53-ddns/main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

deps:
	$(GOGET)
