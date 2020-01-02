FROM golang:latest
RUN mkdir /app
ADD apps/report/* /app/
ADD go.mod /app/
ADD go.sum /app/
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]
