FROM alpine:3.8

RUN mkdir -p /usr/share/ui

WORKDIR /usr/share/ui

COPY statics ./statics

ADD build/ui /usr/share/ui/ui

EXPOSE 9000

ENTRYPOINT /usr/share/ui/ui
