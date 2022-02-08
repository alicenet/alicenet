package MadNet

//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/p2p.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/bootnode.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/aobjs.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/cobjs.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/localstatetypes.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/localstate.proto

//go:generate go build -o=mngen ./cmd/mngen
//go:generate ./mngen -i=./proto/p2p.proto -o=proto -p=proto
//go:generate ./mngen -i=./proto/localstate.proto -o=proto -p=proto
//go:generate sh -c "gofmt -w proto/*_mngen*.go"
//go:generate rm mngen

//go:generate protoc --grpc-gateway_out=:proto/ --proto_path=proto/ proto/localstate.proto
//go:generate protoc --openapiv2_out=:./localrpc/swagger --openapiv2_opt logtostderr=true --proto_path=proto/ proto/localstate.proto

//go:generate go-bindata-assetfs -pkg localrpc -prefix localrpc/swagger/ -o ./localrpc/swagger-bindata/bindata.go localrpc/swagger/...
