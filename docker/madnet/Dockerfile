FROM golang:1.18 AS builder
WORKDIR /madnet
COPY . /madnet
RUN make build

FROM alpine:latest AS base
RUN apk --no-cache add ca-certificates gcompat
WORKDIR /madnet
COPY --from=builder /madnet/madnet ./
CMD ["/madnet/madnet"]

FROM base AS devel
COPY ./snapshot.zip /tmp
RUN unzip -o /tmp/snapshot.zip -d /root && \
    rm /tmp/snapshot.zip