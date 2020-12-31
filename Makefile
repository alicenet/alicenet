GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=madnet
RACE_DETECTOR=madrace

all: test build

build: 
	$(GOCMD) build -o $(BINARY_NAME) ./cmd/main.go

race:
	$(GOCMD) build -o $(RACE_DETECTOR) -race ./cmd/main.go

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
