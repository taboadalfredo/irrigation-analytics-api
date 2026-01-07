FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
