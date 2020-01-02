#!/bin/sh

. config.sh

DIR=$(pwd)

ls -d */ | while read dir
do
		echo $dir
		cd $DIR/$dir
		cat k8s.yaml | sed -e "s/{{PROJECT_ID}}/${PROJECT_ID}/g" -e "s/{{VERSION}}/$VERSION/g" | kubectl --namespace event-example apply --filename -
done
