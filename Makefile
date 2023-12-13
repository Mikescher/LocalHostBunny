DOCKER_REPO=registry.blackforestbytes.com
DOCKER_NAME=mikescher/locbunny

NAMESPACE=$(shell git rev-parse --abbrev-ref HEAD)
HASH=$(shell git rev-parse HEAD)

SWAGGO_VERSION=v1.8.12
SWAGGO=github.com/swaggo/swag/cmd/swag@$(SWAGGO_VERSION)

.PHONY: fmt swagger clean run build-docker run-docker-local inspect-docker push-docker lint

build: enums ids swagger fmt
	mkdir -p _build
	rm -f ./_build/bunny_backend
	go build -v -buildvcs=false -o _build/bunny_backend ./cmd/server

enums:
	go generate models/enums.go

ids:
	go generate models/ids.go

run-systemd: build
	mkdir -p .run-data
	CONF_NS="production" _build/bunny_backend

run-test: build
	mkdir -p .run-data
	sudo BUNNY_LIVERELOAD="$(shell pwd)/webassets" BUNNY_PORT="4004" CONF_NS="local-host" _build/bunny_backend

gow:
	# go install github.com/mitranim/gow@latest
	gow -c run blackforestbytes.com/locbunny/cmd/server

dgi:
	[ ! -f "DOCKER_GIT_INFO" ] || rm DOCKER_GIT_INFO
	echo -n "VCSTYPE="     >> DOCKER_GIT_INFO ; echo "git"                         >> DOCKER_GIT_INFO
	echo -n "BRANCH="      >> DOCKER_GIT_INFO ; git rev-parse --abbrev-ref HEAD    >> DOCKER_GIT_INFO
	echo -n "HASH="        >> DOCKER_GIT_INFO ; git rev-parse              HEAD    >> DOCKER_GIT_INFO
	echo -n "COMMITTIME="  >> DOCKER_GIT_INFO ; git log -1 --format=%cd --date=iso >> DOCKER_GIT_INFO
	echo -n "REMOTE="      >> DOCKER_GIT_INFO ; git config --get remote.origin.url >> DOCKER_GIT_INFO

docker: dgi
	docker build \
            -t "$(DOCKER_NAME):$(HASH)" \
            -t "$(DOCKER_NAME):$(NAMESPACE)-latest" \
            -t "$(DOCKER_NAME):latest" \
            -t "$(DOCKER_REPO)/$(DOCKER_NAME):$(HASH)" \
            -t "$(DOCKER_REPO)/$(DOCKER_NAME):$(NAMESPACE)-latest" \
            -t "$(DOCKER_REPO)/$(DOCKER_NAME):latest" \
            .

swagger-setup:
	mkdir -p ".swaggobin"
	[ -f ".swaggobin/swag_$(SWAGGO_VERSION)" ] || { GOBIN=/tmp/_swaggo go install $(SWAGGO); cp "/tmp/_swaggo/swag" ".swaggobin/swag_$(SWAGGO_VERSION)"; rm -rf "/tmp/_swaggo"; }

swagger: swagger-setup
	".swaggobin/swag_$(SWAGGO_VERSION)" init -generalInfo ./api/router.go --propertyStrategy camelcase --output ./swagger/ --outputTypes "json,yaml" --overridesFile override.swag

run-docker-local: docker
	mkdir -p .run-data
	docker run --rm \
	           --init \
	           --env "CONF_NS=local-docker" \
			   --volume "$(shell pwd)/.run-data/docker-local:/data" \
			   --publish "8080:80" \
			   $(DOCKER_NAME):latest

inspect-docker: docker
	mkdir -p .run-data
	docker run -ti \
	           --rm \
	           --volume "$(shell pwd)/.run-data/docker-inspect:/data" \
	           $(DOCKER_NAME):latest \
	           bash

push-docker:
	docker image push "$(DOCKER_REPO)/$(DOCKER_NAME):$(HASH)"
	docker image push "$(DOCKER_REPO)/$(DOCKER_NAME):$(NAMESPACE)-latest"
	docker image push "$(DOCKER_REPO)/$(DOCKER_NAME):latest"

clean:
	rm -rf _build/*
	rm -rf .run-data/*
	git clean -fdx
	! which go 2>&1 >> /dev/null || go clean
	! which go 2>&1 >> /dev/null || go clean -testcache

fmt: swagger-setup
	go fmt ./...
	".swaggobin/swag_$(SWAGGO_VERSION)" fmt

.PHONY: test
test:
	go test ./test/...

lint:
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.53.2
	golangci-lint run ./...
