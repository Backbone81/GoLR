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

.PHONY: run
run: prepare
	go run ./ "4 + 5 * 3"

.PHONY: test
test: prepare
	go test ./...

.PHONY: clean
clean:
	rm -f calculator
