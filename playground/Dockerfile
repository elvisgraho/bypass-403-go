FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/403-playground .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /out/403-playground /app/403-playground

EXPOSE 8080

CMD ["/app/403-playground"]
