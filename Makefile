BINARY_NAME=madnet
RACE_DETECTOR=madrace
YELLOW=\033[0;33;1m
NOCOL=\033[31;0m

build:
	go build -o $(BINARY_NAME) ./cmd/main.go;

race:
	go build -o $(RACE_DETECTOR) -race ./cmd/main.go;

generate: generate-bridge generate-go

generate-bridge:
	export MSYS_NO_PATHCONV=1 &&\
	export PASS_PERMVARS=1 &&\
	mkdir bridge/node_modules 2>/dev/null || true &&\
	docker/update-container.sh docker/generate-bridge/Dockerfile madnet-generate-bridge "-v $$PWD/bridge:/app -v /app/node_modules/" &&\
	docker start -ia madnet-generate-bridge

generate-go:
	export MSYS_NO_PATHCONV=1 &&\
	export PASS_PERMVARS=1 &&\
	./docker/update-container.sh docker/generate-go/Dockerfile madnet-generate-go "-v $$PWD:/app -v /app/bridge -v /app/.git" &&\
	docker start -ia madnet-generate-go

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(RACE_DETECTOR)
	rm -rf localrpc/swagger-bindata/bindata.go localrpc/swagger/localstate.swagger.json **/*.pb.go **/*.capnp.go **/*.pb.gw.go **/*_mngen.go **/*_mngen_test.go bridge/artifacts bridge/bindings bridge/cache bridge/typechain-types bridge/node_modules
	
	docker container rm -vf madnet-generate-go madnet-generate-bridge 2> /dev/null
	docker image rm -f madnet-generate-go madnet-generate-bridge 2> /dev/null
	docker builder prune -f