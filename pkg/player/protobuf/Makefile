
all: go

go:
	protoc --go_out=. --go-grpc_out=. player.proto
	mv github.com/xaionaro-go/player/pkg/player/protobuf/go .
	rm -rf github.com

cpp:
	mkdir -p cpp
	protoc --grpc_out=./cpp/ --plugin=protoc-gen-grpc="$(shell which grpc_cpp_plugin)" player.proto
	protoc --cpp_out=./cpp/ player.proto
