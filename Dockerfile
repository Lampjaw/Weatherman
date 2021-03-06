FROM golang:1.12.7-alpine AS build-env

RUN apk add -U --no-cache build-base git

RUN mkdir /build
RUN mkdir /bot
WORKDIR /bot

ADD . .

RUN go get -d ./... && \
    go build -v -o /build/bot ./cmd/bot

FROM alpine:latest

RUN apk add -U --no-cache iputils ca-certificates tzdata

COPY --from=build-env /build /bin

CMD [ "/bin/bot" ]