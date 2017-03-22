#!/bin/sh
ASTEXPORT_MOD="$HOME/pyenv/versions/3.6.0/envs/gogen/lib/python3.6/site-packages/pydetector/astexport.py"
python3 $ASTEXPORT_MOD sources/assert_constant.py > native/assert_constant.py.json
python3 $ASTEXPORT_MOD sources/astexport.py > native/astexport.py.json
python3 $ASTEXPORT_MOD sources/augassign.py > native/augassign.py.json
python3 $ASTEXPORT_MOD sources/classdef.py > native/classdef.py.json
python3 $ASTEXPORT_MOD sources/comments.py > native/comments.py.json
python3 $ASTEXPORT_MOD sources/comments.py.orig > native/py.py.json
python3 $ASTEXPORT_MOD sources/comprehension.py > native/comprehension.py.json
python3 $ASTEXPORT_MOD sources/except.py > native/except.py.json
python3 $ASTEXPORT_MOD sources/functiondef.py > native/functiondef.py.json
python3 $ASTEXPORT_MOD sources/functions.py > native/functions.py.json
python3 $ASTEXPORT_MOD sources/hello.py > native/hello.py.json
python3 $ASTEXPORT_MOD sources/ifexpression.py > native/ifexpression.py.json
python3 $ASTEXPORT_MOD sources/import.py > native/import.py.json
python3 $ASTEXPORT_MOD sources/imports.py > native/imports.py.json
python3 $ASTEXPORT_MOD sources/line_comment.py > native/line_comment.py.json
python3 $ASTEXPORT_MOD sources/literals_assign.py > native/literals_assign.py.json
python3 $ASTEXPORT_MOD sources/loop_if.py > native/loop_if.py.json
python3 $ASTEXPORT_MOD sources/pass.py > native/pass.py.json
python3 $ASTEXPORT_MOD sources/sameline.py > native/sameline.py.json
python3 $ASTEXPORT_MOD sources/while.py > native/while.py.json
python3 $ASTEXPORT_MOD sources/with.py > native/with.py.json
