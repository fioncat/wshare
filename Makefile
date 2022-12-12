.PHONY: build-server
build-server:
	go build -o ./bin/wshare-server ./cmd/wshare-server

.PHONY: build
build:
	go build -o ./bin/wshare ./cmd/wshare

.PHONY: install
install:
	go install ./cmd/wshare
