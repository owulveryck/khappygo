VERSION=v0.7.6
TARGETS=yolo emotion image-extractor pigo
TARGETSK8S=yolo/k8s.yaml emotion/k8s.yaml image-extractor/k8s.yaml pigo/k8s.yaml
.PHONY: $(TARGETS) $(TARGETSK8S)


all: $(TARGETS)
deploy: $(TARGETSK8S)

$(TARGETS):
	@echo $@
	@go build -o /dev/null $@/*.go
	gcloud builds submit $@ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$@:$(VERSION) 2>&1 | grep -A2 "STATUS"

$(TARGETSK8S):
	@echo $@
	@cat $@ | sed -e "s/{{PROJECT_ID}}/${PROJECT_ID}/g" -e "s/{{VERSION}}/$(VERSION)/g" | kubectl --namespace event-example apply --filename -

