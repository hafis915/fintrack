#!/usr/bin/env bash
set -euo pipefail
cmd=${1:-up}
migrate -path database/migrations -database "${DATABASE_URL}" "$cmd"
