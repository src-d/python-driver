#!/bin/sh
time sh -c './sendmsg.py /usr/lib/python2.7/*.py | docker run -i --name python_driver --rm python_driver_alpine:latest > /dev/null'
