FROM python:3.6-alpine
MAINTAINER source{d}

RUN apk add --update python py-pip git && rm -rf /var/cache/apk/*
RUN pip3 install msgpack-python six \
        git+https://github.com/juanjux/python-pydetector.git
RUN pip2 install msgpack-python six \
        git+https://github.com/juanjux/python-pydetector.git
ADD bin /bin

CMD ["python3",  "bin/python_driver.py"]
