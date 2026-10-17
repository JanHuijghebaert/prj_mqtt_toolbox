#!/usr/bin/env bash

set -euo pipefail

mosquitto_sub -h localhost -u <username> -P <password> -t '$SYS/#' -v
