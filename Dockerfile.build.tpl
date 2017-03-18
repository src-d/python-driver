FROM python:3.6-alpine

RUN mkdir -p /opt/driver/src && \
    adduser $BUILD_USER -u $BUILD_UID -D -h /opt/driver/src

# XXX remove vim
RUN apk add --no-cache --update python py-pip git build-base bash vim
RUN pip3 install six git+https://github.com/juanjux/python-pydetector.git \
    git+https://github.com/python/mypy.git@0bb2d1680e8b9522108b38d203cb73021a617e64#egg=mypy-lang

RUN pip2 install six git+https://github.com/juanjux/python-pydetector.git
ADD native/python_package /tmp/python_driver
RUN pip3 install /tmp/python_driver
RUN yes|rm -rf /tmp/python_driver

WORKDIR /opt/driver/src
