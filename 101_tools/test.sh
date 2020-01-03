
gsutil cp ../21_apps/testdata/meme.jpg gs://khappygo
kubectl --namespace event-example exec curl -- curl -v "http://default-broker.event-example.svc.cluster.local" \
  -X POST \
  -H "Ce-Id: test" \
  -H "Ce-Specversion: 0.3" \
  -H "Ce-Type: image.png" \
  -H "Ce-Source: not-sendoff" \
  -H "Content-Type: text/plain" \
  -d 'gs://khappygo/meme.jpg'

kubectl --namespace event-example  logs --selector app=yolo
