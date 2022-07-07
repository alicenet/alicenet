SHELL=/bin/bash

BINARY_NAME=alicenet
RACE_DETECTOR=alicerace

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
	buf breaking --against '.git#branch=main'

.PHONY: format
format:
	buf format -w
	go mod tidy

.PHONY: generate
generate: generate-bridge generate-go

.PHONY: generate-bridge
generate-bridge: init
	find . -iname \*.capnp.go \
		   -exec rm -rf {} \;
	cd bridge && npm run build

.PHONY: generate-go
generate-go: init
	find . -iname \*.pb.go \
	    -o -iname \*.pb.gw.go \
	    -o -iname \*_mngen.go \
		-o -iname \*_mngen_test.go \
		-o -iname \*.swagger.json \
		-o -iname \*.mockgen.go \
		-exec rm -rf {} \;
	go generate -tags tools ./...

.PHONY: clean
clean:
	go clean
	rm -f $(BINARY_NAME) $(RACE_DETECTOR)

.PHONY: setup
setup:
	go mod download
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
	cd bridge && npm install
