SHELL=/bin/bash

BINARY_NAME=madnet
RACE_DETECTOR=madrace

.PHONY: init
init:
	./scripts/base-scripts/init-githooks.sh

.PHONY: build
build: init
	go build -o $(BINARY_NAME) ./cmd/main.go;

.PHONY: race
race:
	go build -o $(RACE_DETECTOR) -race ./cmd/main.go;

.PHONY: lint
lint:
	golangci-lint run
	buf lint
	buf breaking --against '.git#branch=candidate'

.PHONY: format
format:
	buf format -w

.PHONY: generate
generate: generate-bridge generate-go

.PHONY: generate-bridge
generate-bridge: init
	find . -iname \*.capnp.go \
	       -o -iname bridge/bindings \
		   -exec rm -rf {} \;
	cd bridge && npm install && npm run build

.PHONY: generate-go
generate-go: init
	find . -iname \*.pb.go \
	    -o -iname \*.pb.gw.go \
	    -o -iname \*_mngen.go \
		-o -iname \*_mngen_test.go \
		-o -iname \*.swagger.json \
		-o -iname \*.mockgen.go \
		-exec rm -rf {} \;
	go generate ./...

.PHONY: clean
clean:
	go clean
	rm -f $(BINARY_NAME) $(RACE_DETECTOR)
  