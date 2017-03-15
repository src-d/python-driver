FROM python:3.6-alpine

RUN mkdir -p /opt/driver/src && \
    adduser $BUILD_USER -u $BUILD_UID -D -h /opt/driver/src

RUN apk add --no-cache --update python py-pip git build-base
RUN pip3 install six git+https://github.com/juanjux/python-pydetector.git \
    git+https://github.com/python/mypy.git@0bb2d1680e8b9522108b38d203cb73021a617e64#egg=mypy-lang

RUN pip2 install six git+https://github.com/juanjux/python-pydetector.git
ADD native/python_package /opt/python_driver
RUN pip3 install /opt/python_driver

WORKDIR /opt/driver/src
