#!/usr/bin/env bash

set -euo pipefail

rm -rf venvmc
python3 -m venv venvmc
source venvmc/bin/activate
pip list
pip install -r requirements.txt
pip list
python mqtt_client.py -h
deactivate
