module github.com/MadBase/MadNet

go 1.15

require (
	// github.com/MadBase/bridge v0.0.0-00010101000000-000000000000
	github.com/MadBase/bridge v0.8.0
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/dgraph-io/ristretto v0.0.2 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/emicklei/proto v1.9.0
	github.com/ethereum/go-ethereum v1.10.8
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.8
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d
	github.com/holiman/uint256 v1.2.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pborman/uuid v0.0.0-20170112150404-1b00554d8222
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tinylib/msgp v1.1.5 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20210414172302-67b91c9a5c03
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	golang.org/x/sys v0.0.0-20210816183151-1e6c022a8912
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	zombiezen.com/go/capnproto2 v2.18.0+incompatible
)

// replace github.com/MadBase/bridge => ../bridge
