#!/bin/bash

cd packer-post-processor-buildkite

# Cleanup
rm -f *_test_out*.tfvars
rm -f crash.log
rm -f packer-post-processor-buildkite


# Install deps
go get


# Testing
go test ./...
if [[ $? != 0 ]]; then
    exit 1
fi


# Linting
go vet ./...
if [[ $? != 0 ]]; then
    exit 1
fi

golint ./...
if [[ $? != 0 ]]; then
    exit 1
fi

gocyclo -over 15 .
if [[ $? != 0 ]]; then
    exit 1
fi


# Formatting
gofmt -s -d -l *.go
if [[ $? != 0 ]]; then
    exit 1
fi


# Build
go build
if [[ $? != 0 ]]; then
    exit 1
fi


# Run
BUILDKITE_AGENT_ENDPOINT=https://agent.buildkite.com/v3 TMPDIR=~/tmp $GOPATH/bin/packer build -var-file=../multi-test-variables.json ../multi-test.json
if [[ $? != 0 ]]; then
    exit 1
fi
