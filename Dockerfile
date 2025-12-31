FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git nodejs npm

COPY go.mod go.sum ./
RUN go mod download

COPY web/package*.json ./web/
RUN cd web && npm ci

COPY . .

RUN cd web && npm run build

RUN CGO_ENABLED=1 GOOS=linux go build -o upgo ./cmd/server

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates sqlite

COPY --from=builder /app/upgo .
COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/config.yaml.example ./config.yaml

RUN mkdir -p /data

EXPOSE 8081

CMD ["./upgo"]
