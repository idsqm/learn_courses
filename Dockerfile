FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .
RUN CGO_ENABLED=0 go build -o /courses ./cmd/courses

FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates

COPY --from=builder /courses /courses
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/migrations /migrations

EXPOSE 8081

ENTRYPOINT ["/courses"]
