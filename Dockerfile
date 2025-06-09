########################  build  ########################
FROM golang:1.23-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app/reviews ./cmd

# goose CLI
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

########################  runtime  ######################
FROM alpine:3.20

# non-root
RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app

# бинарь, goose, миграции
COPY --from=build /app/reviews .
COPY --from=build /go/bin/goose .
COPY migrations ./migrations

# дефолтные переменные
ENV POSTGRES_DSN="postgres://postgres:postgres@postgres:5432/reviews-service?sslmode=disable"
ENV OLLAMA_URL="http://ollama:11434"
ENV OLLAMA_MODEL="qwen:7b-chat"

EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./goose -dir ./migrations postgres \"$POSTGRES_DSN\" up && ./reviews"]
