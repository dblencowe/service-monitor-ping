FROM golang:1.18

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o ./service-monitor-ping

FROM alpine:3.13
WORKDIR /app
LABEL org.opencontainers.image.source https://github.com/dblencowe/service-monitor-ping
ENV CGO_ENABLED=0

RUN apk update && apk add tzdata
COPY --from=0 "/app/service-monitor-ping" service-monitor-ping

CMD [ "/app/service-monitor-ping" ]
