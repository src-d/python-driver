#!/bin/sh
ASTEXPORT_MOD="$HOME/pyenv/versions/3.6.0/envs/gogen/lib/python3.6/site-packages/pydetector/astexport.py"
python3 $ASTEXPORT_MOD python/hello.py > python_hello.json
python3 $ASTEXPORT_MOD python/comments.py > python_comments.json
python3 $ASTEXPORT_MOD python/sameline.py > python_sameline.json
python3 $ASTEXPORT_MOD python/imports.py > python_imports.json
python3 $ASTEXPORT_MOD python/astexport.py > python_complex.json
