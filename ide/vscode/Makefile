.PHONY: all
all: build

# Installs dependencies and runs the linter. Other targets depend on this so that a fresh
# checkout always has its node_modules in place and lint-clean sources before building,
# testing, or packaging.
.PHONY: prepare
prepare:
	npm install
	npm run lint

# Type-checks and produces the bundled extension in dist/.
.PHONY: build
build: prepare
	npm run compile

.PHONY: test
test: prepare
	npm test

# Produces an installable extension archive (.vsix) in the current directory. Install it via
# VSCode: Extensions view | "..." menu | Install from VSIX...
.PHONY: package
package: prepare
	npx vsce package

# Launches a fresh VSCode instance with the extension loaded for manual testing (the
# "Extension Development Host").
.PHONY: run
run: prepare
	npm run compile
	code --extensionDevelopmentPath=$(CURDIR)

.PHONY: clean
clean:
	rm -rf dist out .vscode-test *.vsix
