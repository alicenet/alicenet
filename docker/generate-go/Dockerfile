FROM golang:1.17.6-alpine3.15

# some binaries used for code generation can be installed through package manager
RUN apk add --no-cache protoc=3.18.1-r1 capnproto=0.8.0-r1 capnproto-dev=0.8.0-r1

# inherit all permissions from host OS
ARG BUILDER_UID
ARG BUILDER_GIDS
ADD docker/apply-permvars-alpine.sh /
RUN /apply-permvars-alpine.sh
USER $BUILDER_UID

WORKDIR /app
ADD --chown=$BUILDER_UID go.mod go.sum /app/

# other binaries used for code generation are installed from go source
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go &&\
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc &&\
    go install github.com/go-bindata/go-bindata/v3/... &&\
    go install github.com/elazarl/go-bindata-assetfs/... &&\
    go install github.com/MadBase/go-capnproto2/v2/capnpc-go &&\
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway &&\
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

CMD go generate -x ./... 