#!/bin/bash
# Note: This must be run from the parent directory, not from tests
coverage run --source=python_driver/,. --omit=sendmsg.py  -m unittest discover
coverage report
