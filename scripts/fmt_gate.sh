#!/bin/sh
# The formatter gate: gofumpt-as-linter does not exist in golangci-lint v2,
# so unformatted files are caught here, before lint. A formatting drift is
# a run failure, not a --fix suggestion. deps/ is excluded: third-party
# source is not written under this policy and moves out with the
# degraded-network workaround.
set -e

out=$(mktemp)
trap 'rm -f "$out"' EXIT

golangci-lint fmt --diff . > "$out" 2>&1 || true

drift=$(grep '^diff' "$out" | grep -vE '^diff deps/' | sed 's/^diff //; s/\.orig .*//' || true)

if [ -n "$drift" ]; then
    echo "formatter gate: files not gofumpt-clean (run golangci-lint fmt):" >&2
    echo "$drift" >&2
    exit 1
fi

echo "formatter gate: clean"
