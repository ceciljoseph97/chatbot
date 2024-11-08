FROM golang:1.21-alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o perichat ./cli/chat
FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/perichat .
COPY ./cli/etc /app/cli/etc
COPY ./cli/config.yaml /app/cli/config.yaml
COPY ./cli/chat/*.gob /app/

EXPOSE 8080
CMD ["./perichat", "-config", "/app/cli/config.yaml", "-c", "/app/PMFuncOverview.gob"]
