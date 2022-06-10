test:
	go test -race -v ./pkg/...

push:
	dagger do push

dep:
	go get -u ./...