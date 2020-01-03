#!/bin/sh

kubectl --namespace event-example apply --filename https://raw.githubusercontent.com/google/knative-gcp/master/config/300-storage.yaml
for i in *.yaml
do
	kubectl apply --namespace event-example --filename $i
done
