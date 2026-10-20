FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o runtir-api cmd/main.go cmd/api.go

FROM alpine:3.19 AS runner

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/runtir-api .
COPY docs/ ./docs/

EXPOSE 8080

CMD ["./runtir-api"]
