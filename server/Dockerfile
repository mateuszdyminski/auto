FROM alpine:3.8

RUN mkdir -p /usr/share/server

WORKDIR /usr/share/server

COPY config/kube.toml ./config/kube.toml

ADD build/server /usr/share/server/server

ENTRYPOINT /usr/share/server/server -config=/usr/share/server/config/kube.toml
