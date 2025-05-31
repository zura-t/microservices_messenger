# Используем bin в текущей директории для установки плагинов protoc
LOCAL_BIN := $(CURDIR)/bin

# Добавляем bin в текущей директории в PATH при запуске protoc
PROTOC = PATH="$$PATH:$(LOCAL_BIN)" protoc

# Путь до protobuf файлов
PROTO_PATH := proto/api

# Путь до сгенеренных .pb.go файлов
PKG_PROTO_PATH := "$(CURDIR)/pkg"

# устанавливаем необходимые плагины
.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# генерация .go файлов с помощью protoc
.protoc-generate:
	protoc --proto_path=$(PROTO_PATH) \
	--go_out=pkg --go_opt paths=source_relative \
	--go-grpc_out=pkg --go-grpc_opt paths=source_relative \
	$(PROTO_PATH)/mailer/*.proto

# go mod tidy
.tidy:
	go mod tidy

# Генерация кода из protobuf
generate: .bin-deps .protoc-generate .tidy

# Билд приложения
build:
	go build -o $(LOCAL_BIN) ./cmd/api-gateway/client
	go build -o $(LOCAL_BIN) ./cmd/api-gateway/server

# Объявляем, что текущие команды не являются файлами и
# интсрументируем Makefile не искать изменения в файловой системе
.PHONY: \
	.bin-deps \
	.protoc-generate \
	.tidy \
	generate \
	build