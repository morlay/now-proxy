FROM golang:1.15 as builder

COPY ./ /go/src/github.com/morlay/now-proxy
WORKDIR /go/src/github.com/morlay/now-proxy

RUN CGO_ENABLED=0 go build -o now-proxy

FROM alpine:latest AS certs

RUN apk add --update --no-cache ca-certificates

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/morlay/now-proxy/now-proxy /go/bin/now-proxy

ENV PORT 80
EXPOSE 80

ENTRYPOINT ["/go/bin/now-proxy"]