VERSION ?= latest
NAME ?= now-proxy
TAG ?= $(VERSION)

build:
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -o ./bin/now-proxy-linux-arm64 ./cmd/now-proxy
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o ./bin/now-proxy-linux-amd64 ./cmd/now-proxy

test:
	go test -race -v ./pkg/...

prepare:
	@echo ::set-output name=image::$(NAME):$(TAG)
	@echo ::set-output name=build_args::VERSION=$(VERSION)

dockerx:
	docker buildx build \
		--push \
		--platform=linux/amd64,linux/arm64 \
		--tag=morlay/now-proxy:latest \
		--file=./cmd/now-proxy/Dockerfile .

dep:
	go get -u ./...