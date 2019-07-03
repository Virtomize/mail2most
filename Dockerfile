From gliderlabs/alpine:latest

Maintainer Carsten Seeger <info@casee.de>

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN update-ca-certificates

RUN mkdir -p /mail2most/conf
WORKDIR /mail2most
ADD mail2most /mail2most/
ADD conf/mail2most.conf /mail2most/conf/mail2most.conf
VOLUME /mail2most/conf
CMD ["./mail2most", "-c", "/mail2most/conf/mail2most.conf"]
