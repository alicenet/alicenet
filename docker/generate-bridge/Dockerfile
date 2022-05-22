# golang helper image used just to compile binaries from go source
FROM golang:1.17.6-alpine3.15 AS go_deps
RUN apk add --no-cache linux-headers=5.10.41-r0 build-base=0.5-r2
RUN go install github.com/ethereum/go-ethereum/cmd/abigen@v1.10.16 &&\
    go install github.com/vburenin/ifacemaker@v1.1.0

# final node image containing binaries compiled by helper image
FROM node:16.14.0-alpine3.15
RUN apk add --no-cache bash=5.1.16-r0
ENV PYTHONUNBUFFERED=1
RUN apk add --update --no-cache python3
RUN ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools
RUN apk add build-base
# install old npm version, because npm >=7 doesnt allow running as root :(
RUN npm i -g npm@6.14.16

# inherit all permissions from host OS
ARG BUILDER_UID
ARG BUILDER_GIDS
ADD docker/apply-permvars-alpine.sh /
RUN /apply-permvars-alpine.sh
USER $BUILDER_UID

WORKDIR /app
ADD --chown=$BUILDER_UID bridge/package.json bridge/package-lock.json /app/
RUN npm ci

COPY --from=go_deps /go/bin/ /usr/local/bin/
CMD npm run build
