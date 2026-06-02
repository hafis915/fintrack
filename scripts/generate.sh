#!/usr/bin/env bash
set -euo pipefail
sqlc generate -f database/sqlc/sqlc.yaml
