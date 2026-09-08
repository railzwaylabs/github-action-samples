# syntax=docker/dockerfile:1
FROM golang:1.25.7-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/sample ./cmd/sample

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app -u 10001 app
COPY --from=build /out/sample /usr/local/bin/sample
COPY db /app/db
USER app
ENV MIGRATIONS_PATH=/app/db
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/sample"]
