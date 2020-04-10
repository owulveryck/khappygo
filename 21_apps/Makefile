VERSION=v0.8.22
TARGETS=ui
TARGETSK8S=yolo/k8s.yaml emotion/k8s.yaml image-extractor/k8s.yaml pigo/k8s.yaml report/k8s.yaml ui/k8s.yaml
.PHONY: $(TARGETS) $(TARGETSK8S) image-extractor pigo emotion report triggers pubsub buckets


all: $(TARGETS) image-extractor pigo emotion report
deploy-anthos: $(TARGETSK8S)

emotion:
	@echo $@
	@go build -o /dev/null $@/*.go
	gcloud builds submit $@ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$@:$(VERSION) 2>&1 | grep -A2 "STATUS"
	gcloud run deploy $@-events --image gcr.io/${PROJECT_ID}/$@:$(VERSION) --memory 1G --allow-unauthenticated \
		--set-env-vars="PUBSUB_TOPIC=emotions,MODEL=gs://aerobic-botany-270918/models/emotions.onnx"

report:
	@echo $@
	@go build -o /dev/null $@/*.go
	gcloud builds submit $@ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$@:$(VERSION) 2>&1 | grep -A2 "STATUS"
	gcloud run deploy $@-events --image gcr.io/${PROJECT_ID}/$@:$(VERSION) --allow-unauthenticated 

pigo:
	@echo $@
	@go build -o /dev/null $@/*.go
	gcloud builds submit $@ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$@:$(VERSION) 2>&1 | grep -A2 "STATUS"
	gcloud run deploy $@-events --image gcr.io/${PROJECT_ID}/$@:$(VERSION) --allow-unauthenticated \
		--set-env-vars="CASCADE_FILE=gs://aerobic-botany-270918/models/facefinder,ANGLE=0.0,IOUTHRESHOLD=0.01,MODEL=gs://aerobic-botany-270918/models/emotions.onnx,PUBSUB_TOPIC=faces"

image-extractor:
	@echo $@
	@go build -o /dev/null $@/*.go
	gcloud builds submit $@ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$@:$(VERSION) 2>&1 | grep -A2 "STATUS"
	gcloud run deploy $@-events --image gcr.io/${PROJECT_ID}/$@:$(VERSION) --allow-unauthenticated \
		--set-env-vars="DEST=gs://aerobic-botany-270918-faces"

buckets:
	gsutil mb -p $(gcloud config get-value project) -l us-central1  gs://aerobic-botany-270918-faces 
	gsutil mb -p $(gcloud config get-value project) -l us-central1  gs://aerobic-botany-270918-input

pubsub:
	gcloud pubsub topics create emotions
	gcloud pubsub topics create faces

triggers: 
	gcloud alpha events triggers create report --target-service report-events --type com.google.cloud.pubsub.topic.publish --parameters topic=emotions
	gcloud alpha events triggers create image-extractor --target-service image-extractor-events --type com.google.cloud.pubsub.topic.publish --parameters topic=faces
	gcloud alpha events triggers create face-finder --target-service pigo-events --type com.google.cloud.auditlog.event --parameters methodName=storage.objects.create,serviceName=storage.googleapis.com
	gcloud alpha events triggers create emotion --target-service emotion-events --type com.google.cloud.auditlog.event --parameters methodName=storage.objects.create,serviceName=storage.googleapis.com

delete-triggers:
	gcloud alpha events triggers delete report --quiet
	gcloud alpha events triggers delete image-extractor --quiet
	gcloud alpha events triggers delete face-finder --quiet
	gcloud alpha events triggers delete emotion --quiet

$(TARGETS):
	@echo $@
	@go build -o /dev/null $@/*.go
	gcloud builds submit $@ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$@:$(VERSION) 2>&1 | grep -A2 "STATUS"
	gcloud run deploy $@-events --image gcr.io/${PROJECT_ID}/$@:$(VERSION) --allow-unauthenticated \
		--set-env-vars="DEST=gs://aerobic-botany-270918/processed,CASCADE_FILE=gs://aerobic-botany-270918/models/facefinder,ANGLE=0.0,IOUTHRESHOLD=0.01,MODEL=gs://aerobic-botany-270918/models/emotions.onnx,PUBSUB_TOPIC=faces"

$(TARGETSK8S):
	@echo $@
	@cat $@ | sed -e "s/{{PROJECT_ID}}/${PROJECT_ID}/g" -e "s/{{VERSION}}/$(VERSION)/g" | kubectl --namespace event-example apply --filename -

