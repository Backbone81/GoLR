#!/usr/bin/env bash

# This script runs GoLR against all known grammars from this repository and outputs scanners and parsers in all possible
# backends into the tmp/testdata/reference folder. That way, the output can be manually inspected and diffed against
# changes in the parser generator itself or the some backend.

set -e

# ================================================================================
# Generate scanners with all backends.
# ================================================================================
SCANNER_INPUTS=(
  "golr:examples/calculator/spec/calculator.golr"
)
SCANNER_BACKENDS=(dot go json yaml)

for SCANNER_INPUT in "${SCANNER_INPUTS[@]}"; do
    FRONTEND="${SCANNER_INPUT%%:*}"
    FRONTEND_FILE_PATH="${SCANNER_INPUT##*:}"
    FILE_NAME="$(basename "${FRONTEND_FILE_PATH}")"
    BACKEND_FILE_DIR="tmp/testdata/reference/${FILE_NAME%.*}"
    mkdir -p "${BACKEND_FILE_DIR}"
    for SCANNER_BACKEND in "${SCANNER_BACKENDS[@]}"; do
        echo "Generating ${BACKEND_FILE_DIR}/scanner.${SCANNER_BACKEND} from ${FRONTEND_FILE_PATH}"
        go run ./cmd/golr scanner \
            --frontend "${FRONTEND}" \
            --frontend-file-path "${FRONTEND_FILE_PATH}" \
            --backend "${SCANNER_BACKEND}" \
            --backend-file-path "${BACKEND_FILE_DIR}/scanner.${SCANNER_BACKEND}"
    done
done

# ================================================================================
# Generate parsers with all backends.
# ================================================================================
PARSER_INPUTS=(
  "bison:testdata/bison-3.8.2.y"
  "bison:testdata/gcc-2.95.3-c.y"
  "bison:testdata/gcc-2.95.3-objc.y"
  "bison:testdata/gcc-3.3.6-cpp.y"
  "bison:testdata/gcc-4.2.4-java.y"
  "bison:testdata/go-1.5.4.y"
  "bison:testdata/postgres-18.4.y"
  "golr:examples/calculator/spec/calculator.golr"
)
PARSER_BACKENDS=(dot go json yaml)

# Produce parser outputs.
for PARSER_INPUT in "${PARSER_INPUTS[@]}"; do
    FRONTEND="${PARSER_INPUT%%:*}"
    FRONTEND_FILE_PATH="${PARSER_INPUT##*:}"
    FILE_NAME="$(basename "${FRONTEND_FILE_PATH}")"
    BACKEND_FILE_DIR="tmp/testdata/reference/${FILE_NAME%.*}"
    mkdir -p "${BACKEND_FILE_DIR}"
    for PARSER_BACKEND in "${PARSER_BACKENDS[@]}"; do
        echo "Generating ${BACKEND_FILE_DIR}/parser.${PARSER_BACKEND} from ${FRONTEND_FILE_PATH}"
        go run ./cmd/golr parser \
            --frontend "${FRONTEND}" \
            --frontend-file-path "${FRONTEND_FILE_PATH}" \
            --backend "${PARSER_BACKEND}" \
            --backend-file-path "${BACKEND_FILE_DIR}/parser.${PARSER_BACKEND}"
    done
done

# ================================================================================
# Convert all GNU Bison grammars.
# ================================================================================
CONVERT_INPUTS=(
  "testdata/bison-3.8.2.y"
  "testdata/gcc-2.95.3-c.y"
  "testdata/gcc-2.95.3-objc.y"
  "testdata/gcc-3.3.6-cpp.y"
  "testdata/gcc-4.2.4-java.y"
  "testdata/go-1.5.4.y"
  "testdata/postgres-18.4.y"
)

# Produce parser outputs.
for CONVERT_INPUT in "${CONVERT_INPUTS[@]}"; do
    FILE_NAME="$(basename "${CONVERT_INPUT}")"
    BACKEND_FILE_DIR="tmp/testdata/reference/${FILE_NAME%.*}"
    mkdir -p "${BACKEND_FILE_DIR}"
    echo "Generating ${BACKEND_FILE_DIR}/grammar.golr from ${CONVERT_INPUT}"
    go run ./cmd/golr convert \
        --input-file-path "${CONVERT_INPUT}" \
        --output-file-path "${BACKEND_FILE_DIR}/grammar.golr"
done
