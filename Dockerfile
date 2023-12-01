

FROM golang:1-alpine AS builder

RUN apk add --no-cache tzdata ca-certificates openssl make git tar coreutils bash

COPY . /buildsrc

RUN cd /buildsrc && make build



FROM alpine:latest

RUN apk add --no-cache tzdata

COPY --from=builder   /buildsrc/_build/bunny_backend  /app/server

RUN mkdir /data

WORKDIR /app

EXPOSE 80

CMD ["/app/server"]
