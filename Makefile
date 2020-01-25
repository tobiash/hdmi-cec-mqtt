GIT_BRANCH := $(shell git rev-parse --always --abbrev-ref HEAD)
DOCKER_IMAGE := tobiasha/hdmi-cec

.PHONY: docker deploy update

docker:
	docker build -t $(DOCKER_IMAGE):$(GIT_BRANCH) .
	docker tag $(DOCKER_IMAGE):$(GIT_BRANCH) $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(GIT_BRANCH)
	docker push $(DOCKER_IMAGE):latest

deploy:
	kubectl apply -k deploy

update:
	kubectl set image hdmi-cec $(DOCKER_IMAGE):$(GIT_BRANCH)
	kubectl rollout restart hdmi-cec
