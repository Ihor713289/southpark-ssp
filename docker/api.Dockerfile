FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api ./
RUN CGO_ENABLED=0 go build -o api .

FROM alpine:latest AS runtime
WORKDIR /root
COPY --from=builder /app/api .
EXPOSE 8080
CMD ["./api"]
