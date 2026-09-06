.PHONY: all
all: build

.PHONY: build
build:
	./gradlew build

# Runs the tests against the version floor, then statically verifies plugin compatibility with the
# currently recommended IDEs via the IntelliJ Plugin Verifier.
.PHONY: test
test:
	./gradlew test
	./gradlew verifyPlugin

# Produces an installable plugin archive in build/distributions/. Install it via
# IntelliJ: Settings | Plugins | gear icon | Install Plugin from Disk.
.PHONY: package
package:
	./gradlew buildPlugin

# Launches a fresh sandbox IDE instance with the plugin loaded for manual testing.
.PHONY: run
run:
	./gradlew runIde

.PHONY: clean
clean:
	./gradlew clean
