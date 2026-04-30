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
	@go test -coverprofile=coverage.out ./pkg/...

cover-summary:
	go tool cover -func=coverage.out

cover-view:
	go tool cover -html=coverage.out

.PHONY: vet
vet:
	go vet ./...

##@ CI

CLUSTER_NAME?=ctrreschk-ci
IMAGE_CI?=dev.kind.local/ci/ctrreschk:latest

.PHONY: ci-build-image
ci-build-image:
	docker build -t $(IMAGE_CI) -f Containerfile .

.PHONY: ci-kind-setup
ci-kind-setup: ci-build-image
	kind create cluster --name $(CLUSTER_NAME) --config hack/ci/kind-ci.yaml
	kind load docker-image --name $(CLUSTER_NAME) $(IMAGE_CI)
	kubectl wait --for=condition=Ready nodes --all --timeout=120s

.PHONY: ci-kind-teardown
ci-kind-teardown:
	kind delete cluster --name $(CLUSTER_NAME)

.PHONY: test-e2e
test-e2e:
	CTRRESCHK_E2E_IMAGE=$(IMAGE_CI) go test -v -count=1 ./test/e2e/...

.PHONY: clean
clean:
	@rm -rf _out
