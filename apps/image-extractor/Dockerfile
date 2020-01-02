FROM golang:latest
RUN mkdir /app
ADD apps/image-extractor/* /app/
ADD go.mod /app/
ADD go.sum /app/
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]
