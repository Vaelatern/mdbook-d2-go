.PHONY: build test

build: go.mod go.sum main.go
	CGO_ENABLED=0 go build

test:
	docker build --target test .

clean:
	rm -f mdbook-d2-go
