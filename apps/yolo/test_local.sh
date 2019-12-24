curl -v "http://localhost:8080/"
  -X POST \
  -H "Ce-Id: myid" \
  -H "Ce-Specversion: 1.0" \
  -H "Ce-Type: image.png" \
  -H "Ce-Source: khappygo-input" \
	-H "Content-Type: text/plain" \
  -d 'file:////test.png'
