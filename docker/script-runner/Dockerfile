# golang helper image used just to compile binaries from go source
FROM golang:1.18-alpine3.15 AS go_deps
RUN apk add --no-cache linux-headers=5.10.41-r0 build-base=0.5-r2
RUN go install github.com/ethereum/go-ethereum/cmd/ethkey@v1.10.8

# final alpine image containing binaries compiled by helper image
FROM alpine:3.15
RUN apk add --no-cache bash=5.1.16-r0 jq=1.6-r1 hexdump=2.37.4-r0
COPY --from=go_deps /go/bin/ /usr/local/bin/

CMD sh