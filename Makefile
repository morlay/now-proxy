up:
	go run ./cmd/now-proxy

test:
	go test -race -v ./pkg/...

dep:
	go get -u ./...

ship:
	wagon do go ship pushx
