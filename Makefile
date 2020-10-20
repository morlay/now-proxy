VERSION ?= latest
NAME ?= now-proxy
TAG ?= $(VERSION)

prepare:
	@echo ::set-output name=image::$(NAME):$(TAG)
	@echo ::set-output name=build_args::VERSION=$(VERSION)

dockerx:
	docker buildx build --platform linux/amd64,linux/arm64 --push -t morlay/now-proxy:latest .