#!/bin/sh
# cd to script dir
cd ${0%/*}

if [ -x ".local/bin/python_driver" ]
then
    export PATH=.local/bin:$PATH
    export PYTHONPATH=.local/lib/python3.6/site-packages:$PYTHONPATH
    python3 .local/bin/python_driver
else
    python3 python_driver
fi
