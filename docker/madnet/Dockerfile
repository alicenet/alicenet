FROM golang:1.18 AS builder
WORKDIR /alicenet
COPY . /alicenet
RUN make build

FROM alpine:latest AS base
RUN apk --no-cache add ca-certificates gcompat
WORKDIR /alicenet
COPY --from=builder /alicenet/alicenet ./
CMD ["/alicenet/alicenet"]

FROM base AS devel
COPY ./snapshot.zip /tmp
RUN unzip -o /tmp/snapshot.zip -d /root && \
    rm /tmp/snapshot.zip
