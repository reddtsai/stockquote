FROM golang:1.14
ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /app
ADD . /app
RUN cd /app && go mod tidy && go build -o app cmd/mining/main.go
CMD ["./app", "-db-host=db"]