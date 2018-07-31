FROM alpine:3.8

RUN mkdir -p /usr/share/indexer

WORKDIR /usr/share/indexer

COPY config/kube.toml ./config/kube.toml

ADD build/indexer /usr/share/indexer

EXPOSE 8080

ENTRYPOINT /usr/share/indexer/indexer -config=/usr/share/indexer/config/kube.toml
