FROM golang:1.24.1-alpine AS base

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

RUN ln -s /go/bin/migrate /usr/local/bin/migrate

WORKDIR /app

COPY go.mod go.sum ./ 

RUN go mod tidy

FROM base AS build

WORKDIR /app

COPY . .

RUN go build -tags prod -o app cmd/app/main.go

FROM base AS production

WORKDIR /app

COPY --from=build /app/app .

COPY --from=build /app/.env ./.env

COPY --from=build /app/migrations ./migrations

COPY --from=base /usr/local/bin/migrate /usr/local/bin/migrate

CMD ["./app"]
