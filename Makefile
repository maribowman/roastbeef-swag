SERVICE		:= screencaster
NAME		:= ghcr.io/fomo-labs/$(SERVICE)
GIT_BRANCH	:= $(shell git rev-parse --abbrev-ref HEAD)
GIT_HASH	:= $(shell git rev-parse --short HEAD)
TAG			:= $(GIT_BRANCH)-$(GIT_HASH)
IMAGE		:= $(NAME):$(TAG)
STAGE		:= local
DOCKER_DATA	:= $(shell cd .. && pwd)/docker-data/$(SERVICE)

### docker
.PHONY: build
build:
	@echo starting build...
	@docker build -q -t $(IMAGE) -t $(NAME):latest .
	@docker image prune -f --filter label=stage=builder > /dev/null

push: build
	@echo pushing images...
	@#az acr login --name registrysandbox
	@docker push $(IMAGE)
	@docker push $(NAME):latest

.PHONE: deploy
deploy: push
	@echo triggering deployment...
	@cd helm && helm upgrade --install $(SERVICE) --values ./values.yaml . --namespace default

up: down
	@DOCKER_DATA=$(DOCKER_DATA) docker-compose -f resources/docker-compose.yaml up -d

down:
	@DOCKER_DATA=$(DOCKER_DATA) docker-compose -f resources/docker-compose.yaml down

.PHONY: service
service: build
	@docker run -d --rm --network=host --name $(SERVICE)_$(TAG) $(NAME):latest > /dev/null

stop:
	@docker stop $$(docker ps -q) > /dev/null

### testing
.PHONY: tests
tests:
	@go test -race ./...

cover:
	@go test -cover ./...

smoke: build
	@docker run -d --rm -p 8800:8800 --name test-runner $(IMAGE) .
	@bash ./test/smoke.sh
	@docker stop test-runner
	@docker rmi $(IMAGE)
