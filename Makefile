GO111MODULE := on
DOCKER_TAG := $(or ${GIT_TAG_NAME}, latest)

all: metal-exporter

.PHONY: metal-exporter
metal-exporter:
	go build -tags netgo -o bin/metal-exporter *.go
	strip bin/metal-exporter

.PHONY: dockerimages
dockerimages:
	docker build -t mwennrich/metal-exporter:${DOCKER_TAG} .

.PHONY: dockerpush
dockerpush:
	docker push mwennrich/metal-exporter:${DOCKER_TAG}

.PHONY: clean
clean:
	rm -f bin/*
