FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
RUN /go/bin/swag init -g cmd/api/main.go -o docs --parseInternal
RUN rm -f docs/docs.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/fastgo ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S -G app app && apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/fastgo /app/fastgo
COPY --from=builder /src/docs /app/docs
COPY --from=builder /src/locales /app/locales

ENV APP_PORT=3005

USER app

EXPOSE 3005

ENTRYPOINT ["/app/fastgo"]
