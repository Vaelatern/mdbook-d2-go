.PHONY: build test

DOCKER?=docker

build: go.mod go.sum main.go
	CGO_ENABLED=0 go build

test:
	$(DOCKER) build --target test .

clean:
	rm -f mdbook-d2-go
