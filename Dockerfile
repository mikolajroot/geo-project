
FROM golang:1.26-alpine AS builder


RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/service ./cmd/${SERVICE_NAME}


FROM alpine:latest


RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app


WORKDIR /app

COPY --from=builder /app/bin/service .
COPY --from=builder /app/db/migrations ./db/migrations
RUN chown -R app:app /app

USER app

ENTRYPOINT ["./service"]