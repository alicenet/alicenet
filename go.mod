module github.com/MadBase/MadNet

go 1.15

require (
	github.com/MadBase/bridge v0.7.0
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/dgraph-io/ristretto v0.0.2 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/emicklei/proto v1.9.0
	github.com/ethereum/go-ethereum v1.10.6
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.8
	github.com/guiguan/caster v0.0.0-20191104051807-3736c4464f38
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d
	github.com/holiman/uint256 v1.2.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/minio/highwayhash v1.0.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/tinylib/msgp v1.1.5 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20210414172302-67b91c9a5c03
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/net v0.0.0-20210410081132-afb366fc7cd1
	golang.org/x/sys v0.0.0-20210420205809-ac73e9fd8988
	google.golang.org/genproto v0.0.0-20200806141610-86f49bd18e98
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	zombiezen.com/go/capnproto2 v2.18.0+incompatible
)

// replace github.com/MadBase/bridge => ../bridge
