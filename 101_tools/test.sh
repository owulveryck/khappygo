
gsutil cp ../21_apps/testdata/100k-ai-faces-1.jpg egs://khappygo
kubectl --namespace event-example exec curl -- curl -v "http://default-broker.event-example.svc.cluster.local" \
  -X POST \
  -H "Ce-Id: test" \
  -H "Ce-Specversion: 0.3" \
  -H "Ce-Type: image.png" \
  -H "Ce-Source: not-sendoff" \
  -H "Content-Type: text/plain" \
  -d 'gs://khappygo/100k-ai-faces-1.jpg'

kubectl --namespace event-example  logs --selector app=yolo
