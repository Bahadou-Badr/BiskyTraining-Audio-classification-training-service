# builder
FROM golang:1.25.4-alpine AS builder
ENV CGO_ENABLED=0
WORKDIR /src

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# copy source
COPY . .

# build
RUN go build -ldflags="-s -w" -o /api-app ./cmd/api

# runtime image
FROM alpine:3.19
# add CA certs so HTTPS works
RUN apk add --no-cache ca-certificates
WORKDIR /app

# copy binary
COPY --from=builder /api-app /app/api-app

# create non-root user (optional)
RUN addgroup -S app && adduser -S -G app app
USER app

EXPOSE 8080
ENTRYPOINT ["/app/api-app"]
