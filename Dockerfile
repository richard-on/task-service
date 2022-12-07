FROM golang:1.19.3-buster as builder

WORKDIR /task

COPY go.* ./
RUN go mod download
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -v -ldflags "-X main.version=0.1.0 -X main.build=`date -u +.%Y%m%d.%H%M%S`" \
    -o run cmd/mail/main.go

FROM alpine:latest

WORKDIR /task

COPY --from=builder /task/run /task/run
COPY --from=builder /task/.env /task/.env

EXPOSE 80

RUN mkdir -p /model/logs && \
    apk update && apk add curl && apk add --no-cache bash && \
    apk add dumb-init
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ./run