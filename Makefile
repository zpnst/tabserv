PROTO_DIR=proto
GEN_DIR=proto/gen

clean:
	@rm ./$(GEN_DIR)/*.go

proto:
	@protoc \
		-I=$(PROTO_DIR) \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/kv.proto

.PHONY: proto clean