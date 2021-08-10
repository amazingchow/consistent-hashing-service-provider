VERSION     := v1.0.0
PROJECT     := github.com/amazingchow/photon-dance-consistent-hashing
SRC         := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
TARGETS     := photon-dance-consistent-hashing
ALL_TARGETS := $(TARGETS)

LDFLAGS += -X "$(PROJECT)/internal/version.Version=$(VERSION)"
LDFLAGS += -X "$(PROJECT)/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "$(PROJECT)/internal/version.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "$(PROJECT)/internal/version.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"

ifeq ($(race), 1)
	BUILD_FLAGS := -race
endif

ifeq ($(debug), 1)
	BUILD_FLAGS += -gcflags=all="-N -l"
endif

compile_pb:
	@docker run --rm -v `pwd`:/defs namely/protoc-all:1.37_0 -o pb -d pb -l go

all: build

build: $(ALL_TARGETS)

$(TARGETS): $(SRC)
ifeq ("$(GOMODULEPATH)", "")
	@echo "no GOMODULEPATH env provided!!!"
	@exit 1
endif
	go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' $(GOMODULEPATH)/$(PROJECT)/cmd/$@

lint:
	@golangci-lint run --skip-dirs=api --deadline=5m

pb-fmt:
	@clang-format -i ./pb/*.proto

test:
	go test -count=1 -v -p 1 $(shell go list ./...)

clean:
	rm -f $(ALL_TARGETS)

.PHONY: all build clean
