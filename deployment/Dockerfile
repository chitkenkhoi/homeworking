FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ../go.mod ../go.sum ./
RUN go mod download

COPY COPY ../ ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/main


EXPOSE 3000


RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD ["/app/main"]