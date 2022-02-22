FROM golang:1.17.6-alpine3.15

RUN apk add --no-cache protoc=3.18.1-r1 capnproto=0.8.0-r1 capnproto-dev=0.8.0-r1

ADD go.mod go.sum /app/
WORKDIR /app

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go &&\
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc &&\
    go install github.com/go-bindata/go-bindata/v3/... &&\
    go install github.com/elazarl/go-bindata-assetfs/... &&\
    go install github.com/MadBase/go-capnproto2/v2/capnpc-go &&\
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway &&\
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
    
CMD go generate -x ./... 