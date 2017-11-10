#FROM jamiehewland/alpine-pypy:3-5.9-slim
FROM alpine:3.6
MAINTAINER source{d}

ARG DEVDEPS=native/dev_deps
ARG CONTAINER_DEVDEPS=/tmp/dev_deps
ARG PYDETECTOR_VER=0.14.2

RUN apk add --no-cache --update python python-dev python3 python3-dev py-pip py2-pip git

ADD native/python_package /tmp/python_driver

ADD ${DEVDEPS} ${CONTAINER_DEVDEPS}
ENV ENV_DEVDEPS=${DEVDEPS}
ENV ENV_PYDETECTOR_VER=${PYDETECTOR_VER}
RUN pip2 install -U ${CONTAINER_DEVDEPS}/python-pydetector || pip2 install pydetector-bblfsh==${PYDETECTOR_VER}
RUN pip3 install -U ${CONTAINER_DEVDEPS}/python-pydetector || pip3 install pydetector-bblfsh==${PYDETECTOR_VER}
RUN yes|rm -rf ${CONTAINER_DEVDEPS}

# pypy3
#RUN /usr/local/bin/pip3 install /tmp/python_driver
# python3
#RUN /usr/bin/pip3 install /tmp/python_driver
RUN pip3 install /tmp/python_driver
RUN yes|rm -rf /tmp/python_driver

ADD build /opt/driver
ENTRYPOINT ["/opt/driver/bin/driver"]
