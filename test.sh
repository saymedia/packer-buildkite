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

# Build for Ubuntu servers
GOARCH=amd64 GOOS=linux go build -o packer-buildkite-linux-amd64/packer-post-processor-buildkite
if [[ $? != 0 ]]; then
    exit 1
fi

# Zip up for a release
zip packer-buildkite-linux-amd64.zip packer-buildkite-linux-amd64/packer-post-processor-buildkite


# Run
$GOPATH/bin/packer build -var-file=../multi-test-variables.json ../multi-test.json
if [[ $? != 0 ]]; then
    exit 1
fi
