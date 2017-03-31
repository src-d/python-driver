FROM python:3.6-alpine
MAINTAINER source{d}

RUN apk add --no-cache --update python py-pip git
RUN pip3 install six git+https://github.com/juanjux/python-pydetector.git
RUN pip2 install six git+https://github.com/juanjux/python-pydetector.git
ADD native/python_package /opt/python_driver
RUN pip3 install /opt/python_driver

CMD /usr/local/bin/python_driver
