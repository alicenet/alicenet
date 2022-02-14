GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=madnet
RACE_DETECTOR=madrace
YELLOW=\033[0;33;1m
NOCOL=\033[31;0m

all: test build

build: 
	$(GOCMD) build -o $(BINARY_NAME) ./cmd/main.go

race:
	$(GOCMD) build -o $(RACE_DETECTOR) -race ./cmd/main.go

generate:
	@ set -eu && \
	DOCKER_BUILDKIT=1 docker build . -f dockerfiles/generate/Dockerfile -t madnet-go-generate; \
	EXISTING=$$(docker ps -a --filter name=madnet-go-generate --format {{.Image}}); \
	\
	if [ "$$EXISTING" = "madnet-go-generate" ]; then \
		IS_RUNNING=$$(docker ps --filter name=madnet-go-generate --format {{.Image}}); \
		if [ "$$IS_RUNNING" != "" ]; then \
			echo "$(YELLOW)  Stopping existing container...  $(NOCOL)"; \
			docker stop -t 0 madnet-go-generate; \
		fi; \
	else \
		if [ "$$EXISTING" != "" ]; then \
			echo "$(YELLOW)  Removing old container...  $(NOCOL)"; \
		  docker rm -f madnet-go-generate; \
		fi; \
		echo "$(YELLOW)  Creating new container...  $(NOCOL)"; \
		docker create --name madnet-go-generate -it -v $$PWD:/app madnet-go-generate; \
	fi; \
	echo "$(YELLOW)  Starting container...  $(NOCOL)"; \
	docker start -ia madnet-go-generate; 

clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
