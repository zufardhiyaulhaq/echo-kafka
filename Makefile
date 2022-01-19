.EXPORT_ALL_VARIABLES:

BIN_DIR := ./bin
OUT_DIR := ./output
$(shell mkdir -p $(BIN_DIR) $(OUT_DIR))

IMAGE_REGISTRY=zufardhiyaulhaq
IMAGE_NAME=$(IMAGE_REGISTRY)/echo-kafka
IMAGE_TAG=$(shell git rev-parse --short HEAD)

CURRENT_DIR=$(shell pwd)
VERSION=$(shell cat ${CURRENT_DIR}/VERSION)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)

STATIC_BUILD?=true

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE}

ifeq (${STATIC_BUILD}, true)
override LDFLAGS += -extldflags "-static"
endif

ifneq (${GIT_TAG},)
IMAGE_TAG=${GIT_TAG}
IMAGE_TRACK=stable
LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
else
IMAGE_TAG?=$(GIT_COMMIT)
IMAGE_TRACK=latest
endif

.PHONY: kafka.up
kafka.up:
	docker-compose --file docker-compose.yaml up --build -d

.PHONY: kafka.down
kafka.down:
	docker-compose --file docker-compose.yaml down

.PHONY: run
run:
	go run .

.PHONY: build
build:
	CGO_ENABLED=0 GO111MODULE=on go build -a -ldflags '${LDFLAGS}' -o ${BIN_DIR}/echo-kafka .

.PHONY: image.build
image.build:
	echo "building container image"
	DOCKER_BUILDKIT=1 docker build \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		--build-arg GITCONFIG=$(GITCONFIG) --build-arg BUILDKIT_INLINE_CACHE=1 .
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(IMAGE_NAME):latest

.PHONY: image.release
image.release:
	echo "pushing container image"
	docker push $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
