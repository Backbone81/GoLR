.PHONY: all
all: build

.PHONY: build
build:
	gcc -std=c99 -Wall -Wextra -Wpedantic -Werror -O2 -o calculator main.c

.PHONY: run
run: build
	./calculator "4 + 5 * 3"

.PHONY: clean
clean:
	rm -f calculator
