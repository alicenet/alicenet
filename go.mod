module github.com/MadBase/MadNet

go 1.15

require (
	github.com/MadBase/bridge v0.0.0-00010101000000-000000000000
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/emicklei/proto v1.9.0
	github.com/ethereum/go-ethereum v1.9.15
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.14.8
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d
	github.com/holiman/uint256 v1.1.1
	github.com/minio/highwayhash v1.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/net v0.0.0-20200930145003-4acb6c075d10
	golang.org/x/sys v0.0.0-20200824131525-c12d262b63d8
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	zombiezen.com/go/capnproto2 v2.18.0+incompatible
)

replace github.com/MadBase/bridge => ../bridge
