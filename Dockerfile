FROM golang:1.19.4-buster as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o capy-content-storage .

FROM alpine:latest
COPY --from=builder /app/capy-content-storage .
RUN mkdir -p /files && mkdir -p /files-removed
CMD ["./capy-content-storage"]