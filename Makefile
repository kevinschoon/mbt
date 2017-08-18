PACKAGES := $(shell go list ./...|grep -v vendor)
DOCKER_IMAGE := mesanine/mbt:latest

.PHONY: docker test

all: ./bin/mbt

test:
	go $@ -v $(PACKAGES)
	go vet $(PACKAGES)

./bin/mbt: test
	mkdir ./bin 2>/dev/null || true
	go build -o ./bin/mbt

docker:
	docker build -t $(DOCKER_IMAGE) .

deploy:
	@docker login -u $$DOCKER_LOGIN $$DOCKER_PASSWORD
	docker push $(DOCKER_IMAGE)
