.PHONY: all
all: build

.PHONY: prepare
prepare:
	npm install --no-fund --no-audit

.PHONY: build
build: prepare
	npm run build

.PHONY: run
run: build
	node main.js "4 + 5 * 3"

.PHONY: clean
clean:
