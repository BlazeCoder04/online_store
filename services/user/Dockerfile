FROM golang:1.24.4 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/app

FROM alpine:3.22.0

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=build /app/server .

COPY --from=build /app/migrations ./migrations

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8081

CMD ["./server"]
