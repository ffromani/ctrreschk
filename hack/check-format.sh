#!/usr/bin/env bash
set -euo pipefail

diff=$(gofmt -s -d cmd pkg internal)
if [ -n "$diff" ]; then
    echo "$diff"
    exit 1
fi
