VERSION=v0.8.18
TARGETS=yolo emotion image-extractor pigo report ui
TARGETSK8S=yolo/k8s.yaml emotion/k8s.yaml image-extractor/k8s.yaml pigo/k8s.yaml report/k8s.yaml ui/k8s.yaml
.PHONY: $(TARGETS) $(TARGETSK8S)


all: $(TARGETS)
deploy-anthos: $(TARGETSK8S)

$(TARGETS):
	@echo $@
	@go build -o /dev/null $@/*.go
	gcloud builds submit $@ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$@:$(VERSION) 2>&1 | grep -A2 "STATUS"
	gcloud run deploy $@-events --image gcr.io/${PROJECT_ID}/$@:$(VERSION) --allow-unauthenticated


$(TARGETSK8S):
	@echo $@
	@cat $@ | sed -e "s/{{PROJECT_ID}}/${PROJECT_ID}/g" -e "s/{{VERSION}}/$(VERSION)/g" | kubectl --namespace event-example apply --filename -

