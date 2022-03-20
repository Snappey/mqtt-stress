FROM golang:1.17-alpine as builder

WORKDIR /app

ENV CGO_ENABLED=1
ENV GOOS=linux

COPY . .
RUN go build -o mqtt-stress

FROM alpine:3.15.0

WORKDIR /app
COPY --from=builder /app/mqtt-stress .

ENTRYPOINT ["/app/mqtt-stress"]