#!/usr/bin/env bash

set -euo pipefail

# This shell script runs every fuzz test of the given packages, one after the other, until its corpus stops growing.
#
# Go's fuzzer cannot do that on its own: -fuzz accepts exactly one fuzz test in exactly one package, and -fuzztime only
# knows a fixed duration. So the script lists the fuzz tests itself and watches the status line the fuzzer prints every
# three seconds. Once "new interesting", the number of inputs which reached new code, has not grown for FUZZ_IDLE_MINUTES,
# the fuzz test is interrupted like with Ctrl-C. The corpus is kept in the Go cache, so the next run continues from it.
#
# Environment:
#   PACKAGE            the packages to fuzz, like for go test (default ./internal/...)
#   FUZZ_TARGET        a regular expression restricting the fuzz tests by name (default: all)
#   FUZZ_IDLE_MINUTES  the minutes without new interesting inputs after which a fuzz test is done (default 10)
#   FUZZ_MAX_MINUTES   the minutes after which a fuzz test is stopped even if still growing (default: no limit)
#   FUZZ_PARALLEL      the number of fuzzing workers (default: GOMAXPROCS, which is every core)
#
# The output of every fuzz test is the plain go test output, also written to tmp/fuzz/ for the script to read and for
# later inspection. A failing input is minimized by Go and written to the testdata/fuzz/
# directory of its package. Committed, it runs as a regression test with every go test.

PACKAGE="${PACKAGE:-./internal/...}"
FUZZ_TARGET="${FUZZ_TARGET:-.}"
FUZZ_IDLE_MINUTES="${FUZZ_IDLE_MINUTES:-10}"
FUZZ_MAX_MINUTES="${FUZZ_MAX_MINUTES:-}"
FUZZ_PARALLEL="${FUZZ_PARALLEL:-}"

# The seconds between two looks at the status line. It is also how fast the end of a fuzz test is noticed.
POLL_SECONDS=1

LOG_DIR=tmp/fuzz
mkdir -p "$LOG_DIR"

# The process group of the running fuzz test. Interrupting that group is exactly what Ctrl-C in a terminal does, which is
# the way Go's fuzzer expects to be stopped: the workers end and the corpus stays in the cache.
running_group=""

stop_running() {
  if [[ -n "$running_group" ]]; then
    kill -INT -- "-$running_group" 2>/dev/null || true
  fi
}

# With job control the fuzz run does not see a Ctrl-C meant for this script, so it is passed on.
on_interrupt() {
  echo
  echo "==> interrupted"
  stop_running
  wait 2>/dev/null || true
  exit 130
}
trap on_interrupt INT TERM

# go test -list prints the matching names of a package followed by the "ok" line naming the package.
targets=()
names=()
while IFS= read -r line; do
  if [[ "$line" =~ ^Fuzz[^[:space:]]*$ ]]; then
    names+=("$line")
  elif [[ "$line" =~ ^ok[[:space:]]+([^[:space:]]+) ]]; then
    for name in "${names[@]}"; do
      targets+=("${BASH_REMATCH[1]} $name")
    done
    names=()
  fi
done < <(go test -list '^Fuzz' "$PACKAGE")

selected=()
for target in "${targets[@]}"; do
  if [[ "${target#* }" =~ $FUZZ_TARGET ]]; then
    selected+=("$target")
  fi
done

if [[ ${#selected[@]} -eq 0 ]]; then
  echo "no fuzz tests in $PACKAGE matching $FUZZ_TARGET"
  exit 1
fi

echo "==> fuzzing ${#selected[@]} fuzz test(s) until idle for $FUZZ_IDLE_MINUTES minute(s)${FUZZ_MAX_MINUTES:+, at most $FUZZ_MAX_MINUTES minute(s) each}"
for target in "${selected[@]}"; do
  echo "    $target"
done

failed=()
for target in "${selected[@]}"; do
  package="${target% *}"
  name="${target#* }"
  log="$LOG_DIR/$(echo "$package" | tr '/' '_').$name.log"

  echo
  echo "==> $package $name (log: $log)"
  # Job control is on only while starting the fuzz test, which gives it a process group of its own without bash printing
  # job notices. tee ignores the interrupt, so it still writes the final lines of go test to the log.
  set -m
  go test -run '^$' -fuzz "^$name\$" ${FUZZ_PARALLEL:+-parallel "$FUZZ_PARALLEL"} "$package" 2>&1 | tee -i "$log" &
  set +m
  running_pid=$!
  running_group="$(jobs -p %+)"

  started=$SECONDS
  last_growth=$SECONDS
  interesting=""
  stopped=""
  while kill -0 "$running_pid" 2>/dev/null; do
    sleep "$POLL_SECONDS"
    status="$(tail -n 20 "$log" | grep '^fuzz: elapsed:' | tail -n 1 || true)"

    # Gathering the baseline coverage of the existing corpus is no idle time.
    if [[ "$status" == *"gathering baseline coverage"* ]]; then
      last_growth=$SECONDS
    elif [[ "$status" =~ new\ interesting:\ ([0-9]+) ]] && [[ "${BASH_REMATCH[1]}" != "$interesting" ]]; then
      interesting="${BASH_REMATCH[1]}"
      last_growth=$SECONDS
    fi

    if ((SECONDS - last_growth >= FUZZ_IDLE_MINUTES * 60)); then
      stopped="idle for $FUZZ_IDLE_MINUTES minute(s)"
    elif [[ -n "$FUZZ_MAX_MINUTES" ]] && ((SECONDS - started >= FUZZ_MAX_MINUTES * 60)); then
      stopped="reached the limit of $FUZZ_MAX_MINUTES minute(s)"
    fi
    if [[ -n "$stopped" ]]; then
      stop_running
      break
    fi
  done

  # With pipefail, this is the exit code of go test and not of tee.
  exit_code=0
  wait "$running_pid" || exit_code=$?
  running_group=""

  # go test exits with 1 after an interrupt even when the fuzz test passed, so a fuzz test stopped by this script has
  # failed only if its output says so.
  if [[ -n "$stopped" ]] && ! grep -q -- '^--- FAIL' "$log"; then
    echo "    done: $stopped"
  elif [[ -z "$stopped" ]] && [[ $exit_code -eq 0 ]]; then
    echo "    done"
  else
    echo "    FAILED, see the output above or $log"
    failed+=("$target")
  fi
done

echo
if [[ ${#failed[@]} -gt 0 ]]; then
  echo "==> ${#failed[@]} of ${#selected[@]} fuzz test(s) failed:"
  for target in "${failed[@]}"; do
    echo "    $target"
  done
  exit 1
fi
echo "==> all ${#selected[@]} fuzz test(s) done"
