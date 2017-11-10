FROM alpine:3.6
#FROM jamiehewland/alpine-pypy:3-5.9-slim

RUN mkdir -p /opt/driver/src && \
    adduser $BUILD_USER -u $BUILD_UID -D -h /opt/driver/src

RUN apk add --no-cache --update python python3 bash make git

WORKDIR /opt/driver/src
