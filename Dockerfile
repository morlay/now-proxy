FROM golang:1.14-alpine as builder

COPY ./ /go/src/github.com/morlay/now-proxy
WORKDIR /go/src/github.com/morlay/now-proxy

RUN go build -o now-proxy

FROM alpine

COPY --from=builder /go/src/github.com/morlay/now-proxy/now-proxy /go/src/github.com/morlay/now-proxy/now-proxy

WORKDIR /go/src/github.com/morlay/now-proxy

ENV PORT 80
EXPOSE 80

ENTRYPOINT ./now-proxy