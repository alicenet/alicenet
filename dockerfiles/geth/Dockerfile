FROM ethereum/client-go:alltools-v1.10.4
WORKDIR /root
COPY ./snapshot.zip /tmp
RUN unzip -d /root /tmp/snapshot.zip && \
    rm /tmp/snapshot.zip
RUN mkdir -p /root/local-geth/geth/ethash && \
    geth makedag 0 /root/.ethash && \
    geth makecache 0 /root/.ethash && \
    geth makedag 36000 /root/.ethash && \
    geth makecache 36000 /root/.ethash
