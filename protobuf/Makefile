# Libs
lib-validate:
	curl https://raw.githubusercontent.com/bufbuild/protovalidate/main/proto/protovalidate/buf/validate/validate.proto --create-dirs -o libs/buf/validate/validate.proto
	curl https://raw.githubusercontent.com/bufbuild/protovalidate/main/proto/protovalidate/buf/validate/expression.proto --create-dirs -o libs/buf/validate/expression.proto
	curl https://raw.githubusercontent.com/bufbuild/protovalidate/main/proto/protovalidate/buf/validate/priv/private.proto --create-dirs -o libs/buf/validate/priv/private.proto

lib-googleapis:
	curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto --create-dirs -o libs/google/api/annotations.proto
	curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto --create-dirs -o libs/google/api/http.proto

libs: lib-validate lib-googleapis

# Generate
auth:
	protoc -I proto -I libs \
		proto/auth/auth.proto \
		--go_out=./gen/go --go_opt=paths=source_relative \
		--go-grpc_out=./gen/go --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=./gen/go --grpc-gateway_opt=paths=source_relative \
		--validate_out="lang=go:./gen/go" --validate_opt=paths=source_relative
