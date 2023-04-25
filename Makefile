# Makefile for releasing podinfo
#
# The release version is controlled from pkg/version

TAG?=4.1.0
NAME:=feishu-chatgpt
DOCKER_REPOSITORY:=blacklee123
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)
GIT_COMMIT:=$(shell git describe --dirty --always)
VERSION:=4.1.0
EXTRA_RUN_ARGS?=

run:
	go run -ldflags "-s -w -X static/pkg/version.REVISION=$(GIT_COMMIT)" *.go \
	--root=./testdata --prefix=/static --port=8000  $(EXTRA_RUN_ARGS)

.PHONY: test
test:
	go test ./... -coverprofile cover.out

build:
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X static/pkg/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/podinfo ./cmd/podinfo/*
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X static/pkg/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/podcli ./cmd/podcli/*

tidy:
	rm -f go.sum; go mod tidy -compat=1.19

vet:
	go vet ./...

fmt:
	gofmt -l -s -w ./
	goimports -l -w ./

build-charts:
	helm lint charts/*
	helm package charts/*

build-container:
	docker buildx build --platform=linux/amd64 -t $(DOCKER_IMAGE_NAME):$(VERSION) .

build-xx:
	docker buildx build \
	--platform=linux/amd64 \
	-t $(DOCKER_IMAGE_NAME):$(VERSION) \
	--load \
	-f Dockerfile .

build-base:
	docker build -f Dockerfile.base -t $(DOCKER_REPOSITORY)/static-base:latest .

push-base: build-base
	docker push $(DOCKER_REPOSITORY)/static-base:latest

test-container:
	@docker rm -f static || true
	@docker run -dp 8000:8000 --name=static $(DOCKER_IMAGE_NAME):$(VERSION)
	@docker ps
	@TOKEN=$$(curl -sd 'test' localhost:9898/token | jq -r .token) && \
	curl -sH "Authorization: Bearer $${TOKEN}" localhost:9898/token/validate | grep test

push-container:
	docker tag $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_IMAGE_NAME):latest
	docker push $(DOCKER_IMAGE_NAME):$(VERSION)
	docker push $(DOCKER_IMAGE_NAME):latest

version-set:
	@next="$(TAG)" && \
	current="$(VERSION)" && \
	/usr/bin/sed -i '' "s/$$current/$$next/g" pkg/version/version.go && \
	echo "Version $$next set in code, deployment, chart and kustomize"

release:
	git tag $(VERSION)
	git push origin $(VERSION)

swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	go get github.com/swaggo/swag/gen@latest
	go get github.com/swaggo/swag/cmd/swag@latest
	cd pkg/api && $$(go env GOPATH)/bin/swag init -g server.go

.PHONY: cue-mod
cue-mod:
	@cd cue && cue get go k8s.io/api/...

.PHONY: cue-gen
cue-gen:
	@cd cue && cue fmt ./... && cue vet --all-errors --concrete ./...
	@cd cue && cue gen