package proto

//go:generate protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. --proto_path=. p2p.proto
//go:generate protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. --proto_path=. bootnode.proto
//go:generate protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. --proto_path=. aobjs.proto
//go:generate protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. --proto_path=. cobjs.proto
//go:generate protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. --proto_path=. localstatetypes.proto
//go:generate protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. --proto_path=. localstate.proto

//go:generate go run ../cmd/mngen -i=./p2p.proto -o=. -p=proto
//go:generate go run ../cmd/mngen -i=./localstate.proto -o=. -p=proto

//go:generate protoc --grpc-gateway_out=:./ --proto_path=./ localstate.proto
//go:generate protoc --openapiv2_out=:../localrpc/swagger --openapiv2_opt logtostderr=true --proto_path=. localstate.proto
