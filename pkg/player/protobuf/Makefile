
all: go

go:
	protoc --go_out=. --go-grpc_out=. player.proto
	mv github.com/xaionaro-go/player/pkg/player/protobuf/go .
	rm -rf github.com

