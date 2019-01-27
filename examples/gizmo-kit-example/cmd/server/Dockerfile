FROM alpine:3.6 as alpine

RUN apk add -U --no-cache ca-certificates

FROM golang:alpine as golang

FROM scratch

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=golang /usr/local/go/lib/time/ /usr/local/go/lib/time/

ADD server /server

ENTRYPOINT ["/server"]
