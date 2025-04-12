serve:
    go tool now-proxy

test:
    go test -v ./pkg/...

dep:
    go get -u ./...

ship:
    wagon do go ship pushx
