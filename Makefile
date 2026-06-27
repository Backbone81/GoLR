PACKAGE ?= ./internal/...

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

.PHONY: build-examples
build-examples: prepare
	$(MAKE) -C examples build

.PHONY: build
build: build-examples
	go build ./cmd/golr

.PHONY: run
run: prepare
	go run ./cmd/golr

.PHONY: test
test: test-examples
	mkdir -p tmp
	rm -rf tmp/coverage
	mkdir -p tmp/coverage
	go test --race -coverpkg=./... -cover $(PACKAGE) -args -test.gocoverdir=$(CURDIR)/tmp/coverage
	@echo
	@echo "========== Correct coverage over all packages =========="
	go tool covdata percent -i=tmp/coverage
	go tool covdata textfmt -i=tmp/coverage -o tmp/cover.out
	go tool cover -html=tmp/cover.out -o tmp/cover.html

.PHONY: test-examples
test-examples: prepare
	$(MAKE) -C examples test

.PHONY: benchmark
benchmark: prepare
	go test -bench=. -benchmem -run=^$$ $(PACKAGE)

.PHONY: clean
clean:
	$(MAKE) -C examples clean
	rm -rf tmp
	rm -f golr

.PHONY: release-test
release-test:
	goreleaser check
	goreleaser healthcheck
	goreleaser build --snapshot --clean
	goreleaser release --snapshot --clean --skip=publish

.PHONY: release
release: release-test
release: export GITHUB_TOKEN ?= unknown
release:
	goreleaser release
