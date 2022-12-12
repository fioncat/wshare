.PHONY: build-daemon
build-daemon:
	go build -o ./bin/wshared ./cmd/wshared

.PHONY: build-server
build-server:
	go build -o ./bin/wshare-server ./cmd/wshare-server

.PHONY: build-cli
build-cli:
	go build -o ./bin/wshare ./cmd/wshare

.PHONY: build
build: build-daemon build-server build-cli
