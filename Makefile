PACKAGE ?= ./...

.PHONY: all
all: build

.PHONY: lint
lint:
	docker run \
		--tty \
		--rm \
		--volume ${PWD}:/app:ro \
		--workdir /app \
		golangci/golangci-lint:v2.11.4 \
		golangci-lint run $(PACKAGE)

.PHONY: generate
generate:
	go run ./examples/bison/spec/export/
	go run ./examples/golang/spec/export/
	go run ./examples/golang/parser/export/

.PHONY: prepare
prepare: generate
	go mod tidy
	go fmt $(PACKAGE)
	go vet $(PACKAGE)

.PHONY: build
build: prepare
	go build ./cmd/golr

.PHONY: run
run: prepare
	go run ./cmd/golr

.PHONY: test
test: prepare
	rm -rf tmp/coverage
	mkdir -p tmp/coverage
	go test --race -coverpkg=./... -cover $(PACKAGE) -args -test.gocoverdir=$(CURDIR)/tmp/coverage
	go tool cover -html tmp/coverprofile.out -o tmp/coverprofile.html
	@echo
	@echo "========== Correct coverage over all packages =========="
	go tool covdata percent -i=tmp/coverage
	go tool covdata textfmt -i=tmp/coverage -o tmp/cover.out
	go tool cover -html=tmp/cover.out -o tmp/cover.html

.PHONY: benchmark
benchmark: prepare
	go test -bench=. -benchmem -run=xxx $(PACKAGE)

.PHONY: clean
clean:
	rm tmp/*
