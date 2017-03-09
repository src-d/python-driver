#!/bin/sh
docker stop python_driver
docker rm python_driver
docker rmi python_driver:latest
cp Dockerfile.ubuntu Dockerfile
docker build --tag python_driver_ubuntu .
