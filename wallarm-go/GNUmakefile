all: test

test:
	go test ./... -v -timeout=30s -parallel=4 -race

vet:
	go vet $(go list ./...)

cover:
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: test vet cover
