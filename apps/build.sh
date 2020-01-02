#!/bin/sh

. config.sh

DIR=$(pwd)

ls -d */ | while read dir
do
	  dir=$(echo $dir | sed 's|/||')
		cd $DIR/$dir
		go build -o /dev/null && gcloud builds submit ../../ --project ${PROJECT_ID} --tag gcr.io/${PROJECT_ID}/$dir:$VERSION
done

