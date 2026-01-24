# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o app ./cmd/

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=build /src/app /app

EXPOSE 8082

USER nonroot:nonroot
ENTRYPOINT ["/app"]
