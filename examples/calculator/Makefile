.PHONY: all
all: build

.PHONY: build
build:
	$(MAKE) -C golang build
	$(MAKE) -C java build
	$(MAKE) -C rust build

.PHONY: test
test:
	$(MAKE) -C golang test

.PHONY: clean
clean:
	$(MAKE) -C golang clean
	$(MAKE) -C java clean
	$(MAKE) -C rust clean
