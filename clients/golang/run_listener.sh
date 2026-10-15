#!/usr/bin/env bash

set -euo pipefail

cd -- "$(dirname -- "$0")"
go run . -topic "#" -username <your_username> -password <your_password>
