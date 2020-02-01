GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
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
	kubectl set image deployment/hdmi-mqtt cec=$(DOCKER_IMAGE):$(GIT_BRANCH)
	kubectl rollout restart deployment/hdmi-mqtt
