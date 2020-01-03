#!/bin/sh

if [ "_$1" == "_" ]
then
	exit
fi
file=$(basename $1)
convert $1 -resize x416 -depth 8 -crop 416x416+0+0 /tmp/$file.jpg
gsutil cp /tmp/$file.jpg gs://khappygo
#gsutil cp $1 gs://khappygo
kubectl --namespace event-example exec curl -- curl -v "http://default-broker.event-example.svc.cluster.local" \
  -X POST \
  -H "Ce-Id: test $file" \
  -H "Ce-Specversion: 0.3" \
  -H "Ce-Type: image.png" \
  -H "Ce-Source: not-sendoff" \
  -H "Content-Type: text/plain" \
  -d "gs://khappygo/$file.jpg"

sleep 2
kubectl --namespace event-example  logs --selector app=emotion
