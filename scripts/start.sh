#!/usr/bin/env bash
# Starts development services

set -eo pipefail

PATH="$(realpath scripts/shims):${PATH}"

go run . &
cd ui && npm start &

wait < <(jobs -p)
