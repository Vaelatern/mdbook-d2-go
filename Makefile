.PHONY: build

build: go.mod go.sum main.go
	CGO_ENABLED=0 go build

clean:
	rm -f mdbook-d2-go
