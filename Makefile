NAME        := k8s-virtual-device-plugin
PROJECTROOT := $(shell pwd)
VERSION     := $(shell cat ${PROJECTROOT}/VERSION)
REVISION    := $(shell git rev-parse --short HEAD)
PACKAGE_PATH := github.com/dtaniwaki/k8s-virtual-device-plugin
IMAGE_PREFIX := dtaniwaki/
IMAGE_NAME := k8s-virtual-device-plugin
IMAGE_TAG   ?= $(VERSION)
GIT_TAG_STATE = $(shell if [ "`git rev-parse --short v$(VERSION)`" = "$REVISION" ]; then echo "clean" ; else echo "dirty"; fi)
GIT_TREE_STATE = $(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
OUTDIR      ?= $(PROJECTROOT)/dist


LDFLAGS := -ldflags="-s -w \
  -X \"main.Version=$(VERSION)\" \
	-X \"main.Revision=$(REVISION)\" \
	-X \"main.GitTagState=$(GIT_TAG_STATE)\" \
	-X \"main.GitTreeState=$(GIT_TREE_STATE)\" \
"


.PHONY: build
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(OUTDIR)/$(NAME)

.PHONY: dev-run
dev-run:
	go run $(LDFLAGS) ./... -no-register $(PROJECTROOT)/device.yaml

.PHONY: install
install:
	go install $(LDFLAGS)

.PHONY: build-image
build-image:
	docker build -t $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_TAG) $(PROJECTROOT)

.PHONY: release-image
release-image: build-image
	docker push $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: minikube-load
minikube-load: IMAGE_TAG = latest
minikube-load: build-image
	docker save $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_TAG) | (eval `minikube docker-env $(IMAGE_PREFIX)$(IMAGE_NAME):$(IMAGE_TAG)` && docker load)

.PHONY: lint
lint:
	golangci-lint run --config golangci.yaml

.PHONY: clean
clean:
	rm -f dist/*
