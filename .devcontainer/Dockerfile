ARG VARIANT=1-bullseye
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

ARG NODE_VERSION="lts/*"
RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c ". /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive && \
    apt-get -y install --no-install-recommends capnproto

RUN su vscode -c "go install github.com/MadBase/go-capnproto2/v2/capnpc-go@v2.18.0-custom-schema.1 && \
    go install github.com/derision-test/go-mockgen/cmd/go-mockgen@v1.3.1 && \
    go install golang.org/x/tools/cmd/goimports@v0.1.10  && \
    go install github.com/ethereum/go-ethereum/cmd/abigen@v1.10.16 && \
    go install github.com/vburenin/ifacemaker@v1.2.0 && \
    go install github.com/ethereum/go-ethereum/cmd/ethkey@v1.10.16 && \
    go install github.com/bufbuild/buf/cmd/buf@v1.5.0"