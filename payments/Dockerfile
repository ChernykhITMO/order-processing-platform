# syntax=docker/dockerfile:1
FROM golang:1.25.1 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux go build -trimpath -ldflags="-s -w" -o app ./cmd/main.go
RUN GOOS=linux go build -trimpath -ldflags="-s -w" -o migrator ./cmd/migrator

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=build /src/app /app

EXPOSE 8082

USER nonroot:nonroot
ENTRYPOINT ["/app"]

FROM gcr.io/distroless/base-debian12 AS migrator
WORKDIR /

COPY --from=build /src/migrator /migrator
COPY migrations /migrations

USER nonroot:nonroot
ENTRYPOINT ["/migrator"]
