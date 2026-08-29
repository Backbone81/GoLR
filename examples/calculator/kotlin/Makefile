.PHONY: all
all: build

.PHONY: build
build:
	kotlinc -jvm-target 17 -Xjdk-release=17 -d classes parser/Scanner.kt parser/Parser.kt Main.kt

.PHONY: run
run: build
	kotlin -classpath classes MainKt "4 + 5 * 3"

.PHONY: clean
clean:
	rm -rf classes
