.PHONY: all
all: build

.PHONY: prepare
prepare:
	go mod tidy
	go fmt ./...
	go vet ./...

.PHONY: build
build: prepare
	go run ../../cmd/golr scanner \
	  --frontend golr \
	  --frontend-file-path spec/golang.golr \
	  --backend go \
	  --backend-file-path parser/scanner.go
	go run ../../cmd/golr parser \
	  --frontend golr \
	  --frontend-file-path spec/golang.golr \
	  --backend go \
	  --backend-file-path parser/parser.go
	go build .

# Target inspect will convert the GoLR grammar into a GNU Bison grammar and call GNU Bison on it to produce counter
# examples for every conflict. This helps with analyzing ambiguities of the grammar.
.PHONY: inspect
inspect:
	go run ../../cmd/golr convert \
	  --input-file-path spec/golang.golr \
	  --output-file-path ../../tmp/golang.y
	bison --warnings=no-other -Werror=conflicts-rr --output=/dev/null --report-file=/dev/null --define=lr.type=ielr --warning=counterexamples ../../tmp/golang.y

.PHONY: test
test: prepare
	go test ./...

.PHONY: benchmark
benchmark: prepare
	go test -bench=. -benchmem -run=^$$ ./...

.PHONY: clean
clean:
	rm -f golang
