.PHONY: all
all: build

.PHONY: build
build:
	python3 -W error -m compileall -q main.py parser

.PHONY: run
run:
	python3 main.py "4 + 5 * 3"

.PHONY: clean
clean:
	rm -rf __pycache__
	rm -rf parser/__pycache__
