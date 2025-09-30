FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o metrics-sidecard cmd/server/main.go

FROM alpine:3.18

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/metrics-sidecard .

RUN chown appuser:appgroup metrics-sidecard

USER appuser

EXPOSE 8080

ENTRYPOINT ["./metrics-sidecard"]

CMD ["-proxy-port", "3000"]