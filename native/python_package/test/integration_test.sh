#!/bin/bash
echo "Running integration test... this usually takes about 30 seconds"
cd test 2> /dev/null
PY3PKGDIR=`python3 -c "import os, sys;sys.stdout.write(os.path.dirname(os.__file__))"`
PY2PKGDIR=`python2 -c "import os, sys;sys.stdout.write(os.path.dirname(os.__file__))"`

./sendmsg.py $PY2PKGDIR/*.py $PY3PKGDIR/*.py|python3 -m python_driver > integration_output.json
CMDOUT=$?
cat integration_output.json | egrep '"status": "(error|fatal)"' > /dev/null
EGREPOUT=$?

if [ $CMDOUT -ne 0 ] || [ $EGREPOUT -eq 0 ]
then
    echo "The integration test failed. Some produced JSON outputs contain fatal" \
        "or error status"
    echo "Batch file parse output status: $CMDOUT"
    echo "Grep for errors output status (1 == no error or fatal matches): $EGREPOUT"
    exit 1
else
    echo "Integration test completed successfully"
    exit 0
fi
