#!/bin/sh
ASTEXPORT_MOD="$HOME/pyenv/versions/3.6.0/envs/gogen/lib/python3.6/site-packages/pydetector/astexport.py"
python3 $ASTEXPORT_MOD sources/hello.py > native/hello.py.json
python3 $ASTEXPORT_MOD sources/comments.py > native/comments.py.json
python3 $ASTEXPORT_MOD sources/sameline.py > native/sameline.py.json
python3 $ASTEXPORT_MOD sources/imports.py > native/imports.py.json
python3 $ASTEXPORT_MOD sources/astexport.py > native/complex.py.json
