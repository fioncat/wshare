.PHONY: build-server
build-server:
	go build -o ./bin/wshare-server ./cmd/wshare-server

.PHONY: build
build:
	CGO_ENABLED=1 go build -o ./bin/wshare ./cmd/wshare

.PHONY: install
install:
	CGO_ENABLED=1 go install ./cmd/wshare
