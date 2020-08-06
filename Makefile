dockerx:
	docker buildx build --platform linux/amd64,linux/arm64 --push -t morlay/now-proxy:latest .