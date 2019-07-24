# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=route53-ddns

all: build

build:
	GO111MODULE=on $(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/route53-ddns/main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

deps:
	$(GOGET)
