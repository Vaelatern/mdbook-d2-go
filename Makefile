.PHONY: build

build: go.mod go.sum main.go
	go build

clean:
	rm -f mdbook-d2-go
