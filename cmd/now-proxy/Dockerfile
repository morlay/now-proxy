FROM --platform=${BUILDPLATFORM} golang:1.16-bullseye as builder

COPY ./ /go/src
WORKDIR /go/src

RUN make build

FROM gcr.io/distroless/static-debian11:latest

ARG TARGETARCH
COPY --from=builder /go/src/bin/now-proxy-linux-${TARGETARCH} /go/bin/now-proxy

ENV PORT 80
EXPOSE 80

ENTRYPOINT ["/go/bin/now-proxy"]