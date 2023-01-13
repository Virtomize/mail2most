From golang:latest as BUILDER

RUN git clone https://github.com/magefile/mage && cd mage && go run bootstrap.go

ADD . /mail2most/

RUN cd /mail2most/ && mage build

From alpine:latest

Maintainer Carsten Seeger <info@casee.de>

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN update-ca-certificates

RUN mkdir -p /mail2most/conf
WORKDIR /mail2most
COPY --from=BUILDER /mail2most/bin/mail2most /mail2most/mail2most
ADD conf/mail2most.conf /mail2most/conf/mail2most.conf
VOLUME /mail2most/conf
CMD ["./mail2most", "-c", "/mail2most/conf/mail2most.conf"]
