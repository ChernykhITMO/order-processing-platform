FROM golang:alpine

ENV GOPATH=/

WORKDIR /cmd/app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o /build ./cmd/app

EXPOSE 8080

CMD ["/build"]



