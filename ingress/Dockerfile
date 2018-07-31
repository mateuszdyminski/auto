FROM alpine:3.8

RUN mkdir -p /usr/share/ingress

WORKDIR /usr/share/ingress

COPY config/kube.toml ./config/kube.toml

COPY data ./data

ADD build/ingress /usr/share/ingress/ingress

ENTRYPOINT /usr/share/ingress/ingress -config=/usr/share/ingress/config/kube.toml
