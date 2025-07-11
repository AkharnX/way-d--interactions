# Dockerfile for way-d-interactions
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o wayd-interactions

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/wayd-interactions .
COPY .env ./
EXPOSE 8082
CMD ["./wayd-interactions"]
