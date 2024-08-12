VERSION:=v0.0.$(shell date +%Y%m%d)
IMAGENAME:=ctrreschk
REGISTRY?=quay.io/fromani
BUILDID?=01
TAG?=$(VERSION)$(BUILDID)
IMAGE:=$(REGISTRY)/$(IMAGENAME):$(TAG)

.PHONY: all
all: build

.PHONY: build
build: binaries

image:
	@podman build -t $(IMAGE) .

binaries: ctrreschk

outdir:
	@mkdir -p _out

ctrreschk: outdir
	CGO_ENABLED=0 go build -v -o _out/ctrreschk cmd/ctrreschk/main.go

test-unit:
	@go test -cover ./pkg/...

.PHONY: vet
vet:
	go vet ./...

.PHONY: clean
clean:
	@rm -rf _out
