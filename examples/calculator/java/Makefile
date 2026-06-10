.PHONY: all
all: build

.PHONY: build
build:
	javac -cp . Main.java parser/Scanner.java

.PHONY: run
run: build
	java Main "4 + 5 * 3"

.PHONY: clean
clean:
	rm -f *.class
	rm -f parser/*.class
