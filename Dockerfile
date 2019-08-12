FROM golang:1.12.7 AS build-env

ENV GO111MODULE=on

WORKDIR /app

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o weatherman ./cmd/weatherman/weatherman.go

CMD [ "./weatherman" ]