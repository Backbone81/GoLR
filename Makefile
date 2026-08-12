PACKAGE ?= ./internal/...

# LANGUAGE restricts test-backends to a single language backend. It is empty by default, which runs all of them.
LANGUAGE ?=

# The languages the harness can test are the runner directories, so adding a language is adding a directory and never
# editing a list. The name of a directory is at once the compose service, the backend the CLI generates with and the
# Ginkgo label which selects the language.
BACKEND_LANGUAGES := $(notdir $(patsubst %/,%,$(wildcard internal/backendtest/runners/*/)))

# BACKEND_LABEL is the Ginkgo label every language backend spec carries. It is what tells the two kinds of test run
# apart: test deselects it because those specs need docker and containers which have not run, and test-backends selects
# it and nothing else because it has just run them. Naming it once is what keeps the selection and the deselection from
# drifting apart.
BACKEND_LABEL := backends

# BACKEND_LABEL_FILTER selects the language backend specs, narrowed to a single language when LANGUAGE is set.
BACKEND_LABEL_FILTER := $(BACKEND_LABEL)$(if $(LANGUAGE), && $(LANGUAGE))

# The shell scripts shellcheck is run over.
SHELL_SCRIPTS := $(sort $(shell find internal scripts -name '*.sh'))

export CGO_ENABLED=0

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
	docker run \
		--tty \
		--rm \
		--volume ${PWD}:/mnt \
		--workdir /mnt \
		koalaman/shellcheck:v0.11.0 \
		$(SHELL_SCRIPTS)

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
	CGO_ENABLED=1 go test --race $(PACKAGE) -args --ginkgo.label-filter='!$(BACKEND_LABEL)'

.PHONY: test-coverage
test-coverage: test-examples
	mkdir -p tmp
	rm -rf tmp/coverage
	mkdir -p tmp/coverage
	CGO_ENABLED=1 go test --race -coverpkg=./... -cover $(PACKAGE) -args -test.gocoverdir=$(CURDIR)/tmp/coverage --ginkgo.label-filter='!$(BACKEND_LABEL)'
	@echo
	@echo "========== Correct coverage over all packages =========="
	go tool covdata percent -i=tmp/coverage
	go tool covdata textfmt -i=tmp/coverage -o tmp/cover.out
	go tool cover -html=tmp/cover.out -o tmp/cover.html

.PHONY: test-examples
test-examples: prepare
	$(MAKE) -C examples test

# test-backends proves that the code every language backend emits behaves like the reference implementation in Go. It
# is separate from test, because it generates, compiles and runs the code in a container per language, which needs
# docker and takes considerably longer than the unit tests. Which of the two runs a spec belongs to is the
# BACKEND_LABEL on the spec: test deselects it, this target selects it and nothing else, so the same package is covered
# by both targets without either running the other's specs.
#
# Producing the traces happens here and not in the suite. The container is a build step: which languages to run is
# exactly the LANGUAGE focus this file already has, and a shell loop expresses it without the suite needing setup nodes
# which run once per group.
#
# It depends on build because every container generates with the golr binary, which is mounted into all of them. That is
# what lets a container do the whole job on its own, so reproducing a failure by hand is one command.
#
# The work directory is emptied first, so a trace can never be left over from an earlier run of a case which no longer
# produces one. That failure would otherwise look like a pass. It is also created here rather than by docker, which
# creates a missing bind mount source as root and leaves the container unable to write to it.
.PHONY: test-backends
test-backends: build
	rm -rf tmp/backendtest
	mkdir -p tmp/backendtest
	for language in $(if $(LANGUAGE),$(LANGUAGE),$(BACKEND_LANGUAGES)); do \
		echo "==> running the corpus through the $$language container"; \
		GOLR_UID=$$(id -u) GOLR_GID=$$(id -g) docker compose \
			--file internal/backendtest/docker-compose.yaml \
			--project-directory . \
			run --rm --no-deps --no-TTY "$$language" || exit 1; \
	done
	go test ./internal/backendtest/... -args --ginkgo.label-filter='$(BACKEND_LABEL_FILTER)'

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

.PHONY: release-notes
release-notes:
	mkdir -p tmp
	scripts/release-notes.sh > tmp/release-notes.md
	@echo
	@echo "========== Release notes =========="
	@cat tmp/release-notes.md

.PHONY: release
release: release-test release-notes
release: export GITHUB_TOKEN ?= unknown
release:
	goreleaser release --clean --release-notes tmp/release-notes.md
