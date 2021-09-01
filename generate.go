package MadNet

//go:generate go build -o=mngen ./cmd/mngen
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/p2p.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/bootnode.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/aobjs.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/cobjs.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/localstatetypes.proto
//go:generate protoc --go_out=proto/ --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=proto/ --proto_path=proto/ proto/localstate.proto
//go:generate ./mngen -i=./proto/p2p.proto -o=proto -p=proto -t=rpc
//go:generate ./mngen -i=./proto/localstate.proto -o=proto -p=proto -t=xservice
//go:generate rm mngen
//go:generate protoc --grpc-gateway_out=:proto/ --proto_path=proto/ proto/localstate.proto
//go:generate protoc --swagger_out=:./localrpc/swagger --swagger_opt logtostderr=true --proto_path=proto/ proto/localstate.proto
//go:generate mv localrpc/swagger/localstate.swagger.json localrpc/swagger/swagger.json
//go:generate go-bindata-assetfs -pkg localrpc -prefix localrpc/swagger/ -o ./localrpc/bindata.go localrpc/swagger/...
//go:generate goimports -w localrpc/bindata.go localrpc/bindata.go
