# Python driver [![Build Status](https://travis-ci.org/bblfsh/python-driver.svg?branch=master)] ![Python version](https://img.shields.io/badge/python%20version-3.6-0.svg)


## Description

This is the Python driver for bblfsh. 

## Status

### Implemented:

- Python version detection

- Decoupled design independent of protocol handlers or input/output buffers

- Improvements over the Python ast module: includes comments and whitespace.

- PEP426 package with dependency information instalable with PIP from github but
  not from Pypi; it will probably never be uploaded there since it doesn't 
  make much sense to use this outside of bblfsh. The Python detection and rich
  AST extraction features are modularized in another package called "pydetector"
  that is already on Pypi. 

- Fully typed used Python 3.5+3.6 type annotations.

- Unittest + integration tests + statical typecheck (using mypy). All are set as
  mandatore in the Travis config.

- Source code fully documented using docstrings.


### TO-DO:

- Adapt to bblfsh SDK and package conventions.

- Byte perfect bidirectional conversion. The generated AST and the whitespace
and comments added group more than one space as one and changes semicolons for
newlines.

- Optimization. There is a lot of room for optimization in the current driver
implementation (the typical Python tricks, pypy-friendliness, etc.)

- Move the rich-AST extractor to its own package (outside of Pydetector) and 
upload to pypi.

- Full unittest coverage of all modules (the NoopExtractor specifically 
lacks many tests).


Installation
------------

Using virtualenv is nice. Using [pyenv](https://github.com/yyuu/pyenv) + 
[pyenv virtualenv](https://github.com/yyuu/pyenv-virtualenv) plugin to avoid
the mess of Linux distributions Python's deployments is god-like.

```bash
# [Install Python 2.7.13 and 3.6.0 in whatever way its needed here]
# (Or change master for the specific tag)
python3.6 -m pip install -U https://github.com/bblfsh/python-driver/archive/master.zip
# Python2.7 is needed to evaluate AST of Python 2.7
python2.7 -m pip install pydetector
```

If you want to develop and/or run the unit/typing/integration tests do this instead:

```bash
# Or whatever you have to do to install these versions on your environment:
pyenv install 3.6.0 
pyenv install 2.7.13 # (or whatever you have to do to install these python
versions)

# Enable Python 2.7 to install pydetector
pyenv global 2.7.13
pip install pydetector

# Download the driver
git clone https://github.com/bblfsh/python-driver python_driver
cd python_driver
pyenv global 3.6.0
pip install -U .
pip install -r requirements.txt
cd test && python3.6 -m unittest discover && sh integration_test.sh
```

License
-------
GPLv3, see [LICENSE](LICENSE)
