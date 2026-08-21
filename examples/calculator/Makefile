.PHONY: all
all: build

.PHONY: build
build:
	$(MAKE) -C cpp build
	$(MAKE) -C golang build
	$(MAKE) -C java build
	$(MAKE) -C javascript build
	$(MAKE) -C python build
	$(MAKE) -C rust build
	$(MAKE) -C typescript build
	GOLR_UID=$$(id -u) GOLR_GID=$$(id -g) docker compose \
		--file $(CURDIR)/../../internal/backendtest/docker-compose.yaml \
		--project-directory $(CURDIR)/../.. \
		run --rm --no-deps --no-TTY \
		--volume $(CURDIR)/csharp:/example \
		--workdir /example \
		--entrypoint dotnet \
		csharp build --nologo --verbosity quiet

.PHONY: test
test:
	$(MAKE) -C golang test

.PHONY: clean
clean:
	$(MAKE) -C cpp clean
	$(MAKE) -C golang clean
	$(MAKE) -C java clean
	$(MAKE) -C javascript clean
	$(MAKE) -C python clean
	$(MAKE) -C rust clean
	$(MAKE) -C typescript clean
	$(MAKE) -C csharp clean
