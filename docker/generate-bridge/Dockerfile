# golang helper image used just to compile binaries from go source
FROM golang:1.17.6-alpine3.15 AS go_deps
RUN apk add --no-cache linux-headers=5.10.41-r0 build-base=0.5-r2
RUN go install github.com/ethereum/go-ethereum/cmd/abigen@v1.10.8

# final node image containing binaries compiled by helper image
FROM node:16.14.0-alpine3.15
RUN apk add --no-cache bash=5.1.16-r0
COPY --from=go_deps /go/bin/ /usr/local/bin/

ADD package.json package-lock.json /app/
WORKDIR /app
RUN npm ci

CMD npm run build