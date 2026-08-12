.PHONY: all
all: build

.PHONY: build
build:
	node --check parser/scanner.js
	node --check parser/parser.js
	node --check main.js

.PHONY: run
run:
	node main.js "4 + 5 * 3"

.PHONY: clean
clean:
