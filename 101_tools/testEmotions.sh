
kubectl --namespace event-example exec curl -- curl -v "http://default-broker.event-example.svc.cluster.local" \
  -X POST \
  -H "Ce-Id: testEmotion" \
  -H "Ce-Specversion: 0.3" \
  -H "Ce-Type: image.partial.png" \
  -H "Ce-Source: test" \
  -H "Extensions: element: face" \
  -H "Content-Type: text/plain" \
  -d 'gs://khappygo/processed/meme_0_face.jpg'

kubectl --namespace event-example  logs --selector app=emotion
