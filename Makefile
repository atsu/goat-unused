GOMOD?=on

IMAGE_SETUP=-v $(shell pwd)/.cpkg:/go/pkg -v $(shell pwd):/src -e BUILD_ENV="$(shell env | grep 'USER\|TRAVIS\|ATSU\|GITHUB')"
IMAGE=$(IMAGE_SETUP) ghcr.io/atsu/gobuilder:latest

build:
	docker run --rm $(IMAGE) sqrl make

cbuild:
	GO111MODULE=$(GOMOD) go build ./...
	GO111MODULE=$(GOMOD) go test ./...

tag:
	git tag $(shell docker run --rm $(IMAGE) sqrl info -v version)

clean:
	chmod -R +w .cpkg
	rm -rf iomkr rpm .cpkg

.PHONY: build cbuild tag clean
