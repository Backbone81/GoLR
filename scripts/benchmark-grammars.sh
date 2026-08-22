#!/usr/bin/env bash

set -eu

# This shell script measures runtime and peak memory of generating an IELR(1) parser for every well-known grammar in
# the testdata folder. In contrast to the Go benchmarks it reports the maximum resident set size of the process, which
# is the actual peak memory of the generator instead of the accumulated allocations including all memory churn.
#
# It expects the golr binary in the root of the project, so run `make build` first. GNU time is used because the shell
# builtin `time` does not report memory consumption.

for grammar in testdata/*.y; do
	echo "=== $(basename "${grammar}" .y)"
	/usr/bin/time --format 'wall %e s, user %U s, sys %S s, peak %M KB' \
		./golr parser \
		--frontend bison \
		--frontend-file-path "${grammar}" \
		--backend null \
		--backend-file-path -
done
