FROM python:3.6-alpine
MAINTAINER source{d}

# XXX remove vim
RUN apk add --no-cache --update python py-pip git build-base bash vim
RUN pip3 install six git+https://github.com/juanjux/python-pydetector.git
RUN pip2 install six git+https://github.com/juanjux/python-pydetector.git

ADD build /opt/driver/bin
ADD native/python_package /tmp/python_driver
RUN pip3 install /tmp/python_driver
RUN yes|rm -rf /tmp/python_driver

CMD /opt/driver/bin/driver
