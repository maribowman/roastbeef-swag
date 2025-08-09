SERVICE		:= roastbeef-swag
NAME		:= ghcr.io/maribowman/$(SERVICE)
GIT_BRANCH	:= $(shell git rev-parse --abbrev-ref HEAD)
GIT_HASH	:= $(shell git rev-parse --short HEAD)
TAG			:= $(GIT_BRANCH)-$(GIT_HASH)
IMAGE		:= $(NAME):$(TAG)
STAGE		:= local
DATA_VOLUME := $(pwd)


### docker
.PHONY: build
build:
	@echo starting build...
	@docker build -q -t $(IMAGE) -t $(NAME):latest .
	@docker image prune -f --filter label=stage=builder >/dev/null

.PHONY: push
push: build
	@echo pushing images...
	@docker push $(IMAGE)
	@docker push $(NAME):latest

.PHONE: deploy
deploy: push
	@echo "Target not implemented"

.PHONY: service
service: build
	@docker run -it --rm -p 8800:8800 -v $(DATA_VOLUME):/data --network=host --name $(SERVICE)_$(TAG) $(NAME):latest $(STAGE)

.PHONY: stop
stop:
	@docker stop $(SERVICE)_$(TAG) > /dev/null
# 	@docker stop $$(docker ps -q) > /dev/null


### testing
.PHONY: run
run:
	@go run .

.PHONY: tests
tests:
	@go test -race ./...

.PHONY: cover
cover:
	@go test -cover ./...

.PHONY: smoke
smoke: build
	@docker run -d --rm -p 8800:8800 --name test-runner $(IMAGE) .
	@bash ./test/smoke.sh
	@docker stop test-runner
	@docker rmi $(IMAGE)
