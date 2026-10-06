# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS build
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 10001 botuser

WORKDIR /app
COPY --from=build /out/bot /app/bot

USER botuser
CMD ["/app/bot"]
