FROM golang:1.17.6-alpine3.15


RUN apk add libmagic file musl-dev libc-dev g++ make fortify-headers patch
RUN apk add linux-headers

RUN apk --no-cache add ca-certificates wget
RUN wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub
RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.28-r0/glibc-2.28-r0.apk
RUN apk add glibc-2.28-r0.apk

RUN wget https://capnproto.org/capnproto-c++-0.9.1.tar.gz &&\
    tar zxf capnproto-c++-0.9.1.tar.gz &&\
    cd capnproto-c++-0.9.1 &&\
    ./configure &&\
    #make -j6 check &&\
    make install