.PHONY: all
all: build

.PHONY: build
build:
	dotnet build --nologo --verbosity quiet

.PHONY: run
run: build
	dotnet run --no-build -- "4 + 5 * 3"

.PHONY: clean
clean:
	rm -rf bin obj
