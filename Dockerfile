FROM golang:1.19.2-buster as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o capybara-content-storage .

FROM alpine:latest
COPY --from=builder /app/capybara-content-storage .
CMD ["./capybara-content-storage"]