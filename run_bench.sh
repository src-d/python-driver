#!/bin/sh

set -e

go build bench_driver.go

echo "Building driver image..."
make build

IMAGE=$(docker images|sed -n 2p|awk '{print $3}')

docker run -d -p 9432:9432 $IMAGE

CONTAINER=$(docker ps|sed -n 2p|awk '{print $1}')

echo "Running with pypy..."
time ./bench_driver python $(python -c "import os; print(os.path.dirname(os.__file__))")/*.py > /dev/null

echo "Running with Python3..."
docker exec -it $CONTAINER sed -i "1s/.*/#!\/usr\/bin\/python3/" /usr/local/bin/python_driver
docker restart $CONTAINER  

time ./bench_driver python $(python -c "import os; print(os.path.dirname(os.__file__))")/*.py > /dev/null

docker stop $CONTAINER
docker rm $CONTAINER
