#!/bin/bash
coverage run --source=/usr/local/lib/python3.6/dist-packages/python_driver,. --omit=sendmsg.py  -m unittest discover
coverage report
