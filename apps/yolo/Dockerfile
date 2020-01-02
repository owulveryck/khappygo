FROM golang:latest
RUN mkdir /app
ADD apps/yolo/* /app/
ADD go.mod /app/
ADD go.sum /app/
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]
