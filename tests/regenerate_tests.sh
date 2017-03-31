#!/bin/bash
# Run this script from the parent directory to regenerate the integration tests with the new UAST numeric codes

LASTIMAGE=dev-`git log -1 --sparse|head -1|cut -f2 -d " "|cut -b1-7`
BASEDIR=$PWD
cd tests/sources
for SOURCE in $(ls *.py)
do
    echo $SOURCE
    cd $BASEDIR
    docker run -it --rm -v $BASEDIR:/code bblfsh/python-driver:$LASTIMAGE /opt/driver/bin/driver parse-native /code/tests/sources/$SOURCE > tests/native/$SOURCE.json
    docker run -it --rm -v $BASEDIR:/code bblfsh/python-driver:$LASTIMAGE /opt/driver/bin/driver parse-uast /code/tests/sources/$SOURCE > tests/uast/$SOURCE.json
done
