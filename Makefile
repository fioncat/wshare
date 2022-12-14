.PHONY: install
install:
	CGO_ENABLED=1 go install ./cmd/wshared

.PHONY: install-config
install-config:
	go install ./cmd/wshare-config

.PHONY: install-server
install-server:
	go install ./cmd/wshare-server
