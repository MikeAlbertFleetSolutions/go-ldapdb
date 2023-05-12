.PHONY: default get codetest test fmt vet

default: fmt vet test

get:
	go get -v ./...

test:
	go test -v -cover

fmt:
	go fmt ./...

vet:
	go vet -all .
