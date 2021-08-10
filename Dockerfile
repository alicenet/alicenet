FROM golang:1.16 AS builder
WORKDIR /madnet
COPY . /madnet
RUN make build

FROM alpine:latest  
RUN apk --no-cache add ca-certificates gcompat
WORKDIR /madnet
COPY --from=builder /madnet/madnet ./
CMD ["/madnet/madnet"]  