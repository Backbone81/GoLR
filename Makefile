PACKAGE ?= ./...

.PHONY: all
all: build

.PHONY: generate
generate:
	scripts/generate.sh

.PHONY: prepare
prepare: generate
	go mod tidy
	go fmt $(PACKAGE)
	go vet $(PACKAGE)
	docker run \
		--tty \
		--rm \
		--volume ${PWD}:/app \
		--workdir /app \
		--user $$(id -u):$$(id -g) \
		--volume $$(go env GOCACHE):/.cache/go-build \
		--env GOCACHE=/.cache/go-build \
		--volume $$(go env GOMODCACHE):/.cache/mod \
		--env GOMODCACHE=/.cache/mod \
		--volume ~/.cache/golangci-lint:/.cache/golangci-lint \
		--env GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		golangci/golangci-lint:v2.12.2 \
		golangci-lint run --fix $(PACKAGE)

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
