# Путь до завендореных protobuf файлов
VENDOR_PROTO_PATH := $(CURDIR)/vendor.protobuf

# vendor
vendor:	.vendor-reset .vendor-protovalidate .vendor-users .vendor-tidy

# delete VENDOR_PROTO_PATH
.vendor-reset:
	rm -rf $(VENDOR_PROTO_PATH)
	mkdir -p $(VENDOR_PROTO_PATH)

# Устанавливаем proto описания buf/validate/validate.proto
.vendor-protovalidate:
	git clone -b main --single-branch --depth=1 --filter=tree:0 \
		https://github.com/bufbuild/protovalidate $(VENDOR_PROTO_PATH)/protovalidate && \
	cd $(VENDOR_PROTO_PATH)/protovalidate
	git checkout
	mv $(VENDOR_PROTO_PATH)/protovalidate/proto/protovalidate/buf $(VENDOR_PROTO_PATH)
	rm -rf $(VENDOR_PROTO_PATH)/protovalidate

# Копируем proto файлы users сервиса
.vendor-users:
	mkdir -p $(VENDOR_PROTO_PATH)/users/api
	cp -f ../users/api/service.proto $(VENDOR_PROTO_PATH)/users/api/

# delete all non .proto files
.vendor-tidy:
	find $(VENDOR_PROTO_PATH) -type f ! -name "*.proto" -delete
	find $(VENDOR_PROTO_PATH) -empty -type d -delete

# Объявляем, что текущие команды не являются файлами и
# интсрументируем Makefile не искать изменения в файловой системе
.PHONY: \
	.vendor-reset \
	.vendor-protovalidate \
	.vendor-users \
	.vendor-tidy \
	vendor
