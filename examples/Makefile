.PHONY: all
all: build

.PHONY: build
build:
	$(MAKE) -C bison build
	$(MAKE) -C calculator build
	$(MAKE) -C golang build

.PHONY: test
test:
	$(MAKE) -C calculator test
	$(MAKE) -C golang test

.PHONY: benchmark
benchmark:
	$(MAKE) -C golang benchmark

.PHONY: clean
clean:
	$(MAKE) -C bison clean
	$(MAKE) -C calculator clean
	$(MAKE) -C golang clean
