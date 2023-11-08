HOSTNAME=registry.terraform.io
NAMESPACE=finalcad
NAME=airbrake
BINARY=terraform-provider-${NAME}
VERSION=9999.99.99
GOOS=linux
GOARCH=amd64
SHELL := /bin/bash

default: install

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

get:
	go get

build: get
	env GOOS=${GOOS} GOARCH=${GOARCH} go build

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${GOOS}_${GOARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${GOOS}_${GOARCH}
	rm .terraform.lock.hcl 2> /dev/null || true

