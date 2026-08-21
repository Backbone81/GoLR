.PHONY: all
all: build

.PHONY: build
build:
	g++ -std=c++17 -Wall -Wextra -Wpedantic -Werror -O2 -o calculator main.cpp

.PHONY: run
run: build
	./calculator "4 + 5 * 3"

.PHONY: clean
clean:
	rm -f calculator
