#!/bin/sh

for i in *.yaml
do
	kubectl apply --namespace event-example --filename $i
done
