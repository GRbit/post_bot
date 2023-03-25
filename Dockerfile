FROM golang:1.20-alpine AS builder

RUN set -ex && \
    apk add --no-progress --no-cache \
    pkgconf git make bash openssl-dev openssh-client-default \
    openldap-dev

# Move to working directory /build
RUN mkdir -p /app
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

# all previous steps could be cached

COPY . .

# Build the application
RUN go get -d -v ./...
RUN go build -o ./app ./cmd/post_bot/main.go

# Build a small image
FROM alpine:3.17.2

RUN set -ex && \
    apk add --no-progress --no-cache \
    pkgconf git make bash openssl-dev openssh-client-default \
    openldap-dev

RUN mkdir -p /app
RUN mkdir -p /opt/app
WORKDIR /app
COPY --from=builder /app/bot /opt/app/bot

EXPOSE 8080

EXPOSE 8443

ENTRYPOINT ["/opt/app/bot"]
