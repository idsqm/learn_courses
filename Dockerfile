FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /courses ./cmd/courses

FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates

COPY --from=builder /courses /courses
COPY --from=builder /app/migrations /migrations

EXPOSE 8081

ENTRYPOINT ["/courses"]
