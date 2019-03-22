# Python driver for [Babelfish](https://github.com/bblfsh/bblfshd) ![Driver Status](https://img.shields.io/badge/status-beta-dbd25c.svg) [![Build Status](https://travis-ci.org/bblfsh/python-driver.svg?branch=master)](https://travis-ci.org/bblfsh/python-driver) ![Native Version](https://img.shields.io/badge/python%20version-3.6.2-aa93ea.svg) ![Go Version](https://img.shields.io/badge/go%20version-1.9-63afbf.svg)

Development Environment
-----------------------

Requirements:
- `docker`
- Go 1.11+
- SDK dependencies _(dep ensure --vendor-only)_

To initialize the build system execute: `go test ./driver`, at the root of the project. This will generate the `Dockerfile` for this driver.

To execute the tests just execute `go run test.go`, this will execute the test over the native and the go components of the driver using Docker.

The build is done executing `go run build.go`. To evaluate the result using a docker container, execute:
`go run build.go test-driver && docker run -it test-driver`.


License
-------

GPLv3, see [LICENSE](LICENSE)



