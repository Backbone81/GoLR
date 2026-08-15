.PHONY: all
all: build

.PHONY: prepare
prepare:
	npm install --no-fund --no-audit
	npm run lint

.PHONY: build
build: prepare
	node --check parser/scanner.js
	node --check parser/parser.js
	node --check main.js

.PHONY: run
run: prepare
	node main.js "4 + 5 * 3"

.PHONY: clean
clean:
