FROM golang:1.27.1-alpine as builder
WORKDIR /app
COPY api ./
RUN go mod download
RUN go build -o api .

FROM alpine:latest as runtime
WORKDIR /root
COPY --from=builder /app/api .
EXPOSE 8080
CMD ["./api"]