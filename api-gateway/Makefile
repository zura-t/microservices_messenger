LOCAL_BIN := $(CURDIR)/bin

PROTOC = PATH="$$PATH:$(LOCAL_BIN)" protoc

PROTO_PATH := proto/api

PKG_PROTO_PATH := "$(CURDIR)/pkg"

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.protoc-generate:
	protoc --proto_path=$(PROTO_PATH) \
	--go_out=pkg --go_opt paths=source_relative \
	--go-grpc_out=pkg --go-grpc_opt paths=source_relative \
	--grpc-gateway_out=pkg --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true \
	$(PROTO_PATH)/api-gateway/*.proto

.tidy:
	go mod tidy

generate: .bin-deps .protoc-generate .tidy

build:
	go build -o $(LOCAL_BIN) ./cmd/api-gateway/client
	go build -o $(LOCAL_BIN) ./cmd/api-gateway/server

.PHONY: \
	.bin-deps \
	.protoc-generate \
	.tidy \
	generate \
	build