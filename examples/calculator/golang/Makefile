.PHONY: all
all: build

.PHONY: prepare
prepare:
	go mod tidy
	go fmt ./...
	go vet ./...

.PHONY: build
build:
	go build .

.PHONY: run
run:
	go run ./ "4 + 5 * 3"

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f golang
