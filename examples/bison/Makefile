.PHONY: all
all: build

.PHONY: prepare
prepare:
	go mod tidy
	go fmt ./...
	go vet ./...

.PHONY: build
build: prepare
	go build .

.PHONY: clean
clean:
	rm -f bison
