FROM alpine:3.6
MAINTAINER source{d}

ARG DEVDEPS=native/dev_deps
ARG CONTAINER_DEVDEPS=/tmp/dev_deps
ARG PYDETECTOR_VER=0.11.2

RUN apk add --no-cache --update python python3 py-pip py2-pip git

ADD build /opt/driver/bin
ADD native/python_package /tmp/python_driver

ADD ${DEVDEPS} ${CONTAINER_DEVDEPS}
ENV ENV_DEVDEPS=${DEVDEPS}
ENV ENV_PYDETECTOR_VER=${PYDETECTOR_VER}
RUN pip2 install -U ${CONTAINER_DEVDEPS}/python-pydetector || pip2 install pydetector-bblfsh==${PYDETECTOR_VER}
RUN pip3 install -U ${CONTAINER_DEVDEPS}/python-pydetector || pip3 install pydetector-bblfsh==${PYDETECTOR_VER}
RUN yes|rm -rf ${CONTAINER_DEVDEPS}

RUN pip3 install /tmp/python_driver
RUN yes|rm -rf /tmp/python_driver

CMD /opt/driver/bin/driver
