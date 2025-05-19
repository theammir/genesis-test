FROM golang:1.24
WORKDIR /app

COPY go.mod go.sum ./

COPY . .
RUN go build -o /usr/local/bin/server -v ./cmd/server

CMD ["server"]
