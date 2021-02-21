SHELL:=/bin/bash

GO=go
PROTOC=protoc

GO_VET_OPTS=-v
GO_TEST_OPTS=-v -race
RM_OPTS=-f

PROTO_SRCS=$(wildcard ./api/protos/*.proto)
PROTO_TYPES_SRCS=$(wildcard ./api/protos/*/*.proto)
GO_PB_DIR=./api/
GO_TYPES_PB_DIR=./api/types/
GO_PB_SRCS=$(addprefix $(GO_PB_DIR),$(patsubst %.proto,%.pb.go,$(notdir $(PROTO_SRCS))))
GO_TYPES_PB_SRCS=$(addprefix $(GO_TYPES_PB_DIR),$(patsubst %.proto,%.pb.go,$(notdir $(PROTO_TYPES_SRCS))))

CMD_DIRS:=$(wildcard cmd/*)
CMDS:=$(subst cmd,bin,$(CMD_DIRS))


%.pb.go: $(PROTO_SRCS) $(PROTO_TYPES_SRCS)
	@$(PROTOC) \
	--go_out=./api/ \
	--go_opt=module=github.com/chez-shanpu/acar/api \
	--go-grpc_out=./api/ \
	--go-grpc_opt=module=github.com/chez-shanpu/acar/api \
	-I ./api/protos/ \
	$^

.SECONDEXPANSION:
#bin/%: $(wildcard cmd/*/*.go) $(wildcard cmd/*/*/*.go) $(wildcard pkg/*/*.go) go.mod bin
bin/%:
	$(GO) build $(GO_BUILD_OPT) -o $@ ./cmd/$*


.PHONY: build
build: $(CMDS)

.PHONY: pb
pb: $(GO_PB_SRCS) $(GO_TYPES_PB_SRCS)

.PHONY: vet
vet:
	$(GO) vet $(GO_VET_OPTS) ./...

.PHONY: test
test: vet
	$(GO) test $(GO_TEST_OPTS) ./...

.PHONY: sudo-test
sudo-test: test
	sudo $(GO) test $(GO_TEST_OPTS) ./...

.PHONY: clean
clean:
	-$(GO) clean
	-rm $(RM_OPTS) bin/*

.PHONY: all
all: sudo-test build

.DEFAULT_GOAL=all
