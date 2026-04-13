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

.PHONY: prepare
prepare:
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
	ginkgo run -p --race --cover --output-dir=tmp $(PACKAGE)
	go tool cover -html tmp/coverprofile.out -o tmp/coverprofile.html

.PHONY: benchmark
benchmark: prepare
	go test -bench=. -benchmem -run=xxx $(PACKAGE)

.PHONY: clean
clean:
	rm tmp/*
