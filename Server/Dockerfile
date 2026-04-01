FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:3.21

RUN apk add --no-cache curl ca-certificates

COPY --from=build /app/server /usr/local/bin/server

ENV COOKEY_HOST=0.0.0.0
ENV COOKEY_PORT=5800
ENV COOKEY_PUBLIC_URL=https://api.cookey.sh

EXPOSE 5800

ENTRYPOINT ["/usr/local/bin/server"]
